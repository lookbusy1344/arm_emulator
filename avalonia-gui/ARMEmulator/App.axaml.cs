using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using ARMEmulator.Views;
using Avalonia;
using Avalonia.Controls;
using Avalonia.Controls.ApplicationLifetimes;
using Avalonia.Markup.Xaml;
using ReactiveUI.Reactive;

namespace ARMEmulator;

public partial class App : Application
{
	private readonly ApplicationMenu applicationMenu = new();

	public override void Initialize()
	{
		AvaloniaXamlLoader.Load(this);
		NativeMenu.SetMenu(this, applicationMenu.Menu);
	}

	public override void OnFrameworkInitializationCompleted()
	{
		if (ApplicationLifetime is IClassicDesktopStyleApplicationLifetime desktop) {
			ComposeMainWindow(desktop, new JsonSettingsStore(JsonSettingsStore.DefaultPath));
		}

		base.OnFrameworkInitializationCompleted();
	}

	/// <summary>
	/// Builds the services and main view model, shows the window, then starts the backend and session in the background.
	/// Everything built here is disposed when the application exits.
	/// </summary>
	private void ComposeMainWindow(IClassicDesktopStyleApplicationLifetime desktop, JsonSettingsStore settingsStore)
	{
		var loaded = settingsStore.Load();
		var settings = loaded.Settings;
		var baseUri = new Uri(settings.BackendUrl);
#pragma warning disable CA2000 // Owned by the application lifetime and disposed in the Exit handler below
		var backend = new BackendManager(settings.BackendUrl);
		var themeDetector = new PlatformThemeDetector();
		var themeService = new ThemeService(themeDetector);
		var http = new HttpClient { BaseAddress = baseUri };
		var ws = new WebSocketClient(BackendEndpoints.WebSocketUri(baseUri).ToString());
#pragma warning restore CA2000
		var fileService = new FileService { RecentFilesLimit = settings.RecentFilesLimit };
		fileService.LoadRecentFiles(settings.RecentFiles);
		var viewModel = new MainWindowViewModel(new ApiClient(http), ws, fileService, settingsStore, new UnsavedChangesPrompt(() => desktop.MainWindow));
		viewModel.ApplySettings(settings);

		_ = viewModel.WhenAnyValue(x => x.Settings)
			.Subscribe(current => themeService.ApplyTheme(current.Theme));

		desktop.MainWindow = new MainWindow(viewModel);
		applicationMenu.Bind(viewModel);
		desktop.Exit += (_, _) => {
			themeService.Dispose();
			fileService.Dispose();
			themeDetector.Dispose();
			viewModel.Dispose();
			backend.Dispose();
			ws.Dispose();
			http.Dispose();
		};

		var startupFile = StartupArguments.FindSourceFile(desktop.Args ?? [], Environment.CurrentDirectory);
		_ = StartAsync(viewModel, backend, loaded.Warning, startupFile);
	}

	/// <summary>Startup failures reach the user through the view model's ErrorMessage; the settings warning shows when there is none.</summary>
	private static async Task StartAsync(MainWindowViewModel viewModel, IBackendManager backend, string? settingsWarning, string? startupFile)
	{
		await viewModel.StartAsync(backend);
		if (startupFile is not null) {
			await viewModel.OpenSourceFileAsync(startupFile);
		}

		viewModel.ErrorMessage ??= settingsWarning;
	}
}
