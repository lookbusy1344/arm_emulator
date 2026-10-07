namespace ARMEmulator.ViewModels;

/// <summary>How far the application has got in reaching its backend.</summary>
public enum ConnectionState
{
	/// <summary>The backend is starting or a session is being created.</summary>
	Connecting,

	/// <summary>A session exists and the event stream is connected.</summary>
	Connected,

	/// <summary>Starting the backend or creating a session failed.</summary>
	Failed
}
