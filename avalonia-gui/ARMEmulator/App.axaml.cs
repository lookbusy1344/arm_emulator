using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using Avalonia;
using Avalonia.Controls.ApplicationLifetimes;
using Avalonia.Markup.Xaml;

namespace ARMEmulator;

public partial class App : Application
{
	public override void Initialize()
	{
		AvaloniaXamlLoader.Load(this);
	}

	public override void OnFrameworkInitializationCompleted()
	{
		if (ApplicationLifetime is IClassicDesktopStyleApplicationLifetime desktop) {
			ComposeMainWindow(desktop, AppSettings.Default);
		}

		base.OnFrameworkInitializationCompleted();
	}

	/// <summary>
	/// Builds the services and main view model, shows the window, then starts the backend and session in the background.
	/// Everything built here is disposed when the application exits.
	/// </summary>
	private static void ComposeMainWindow(IClassicDesktopStyleApplicationLifetime desktop, AppSettings settings)
	{
		var baseUri = new Uri(settings.BackendUrl);
#pragma warning disable CA2000 // Owned by the application lifetime and disposed in the Exit handler below
		var backend = new BackendManager(settings.BackendUrl);
		var http = new HttpClient { BaseAddress = baseUri };
		var ws = new WebSocketClient(BackendEndpoints.WebSocketUri(baseUri).ToString());
#pragma warning restore CA2000
		var viewModel = new MainWindowViewModel(new ApiClient(http), ws, new FileService());

		desktop.MainWindow = new MainWindow(viewModel);
		desktop.Exit += (_, _) => {
			viewModel.Dispose();
			ws.Dispose();
			http.Dispose();
			backend.Dispose();
		};

		// StartAsync reports failures through the view model's ErrorMessage
		_ = viewModel.StartAsync(backend);
	}
}
