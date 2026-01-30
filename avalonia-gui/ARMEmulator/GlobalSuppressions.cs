using System.Diagnostics.CodeAnalysis;

// Domain exceptions use primary constructors with specific parameters - don't need standard exception constructors
[assembly: SuppressMessage("Design", "RCS1194:Implement exception constructors", Justification = "Domain exceptions use primary constructors", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Services")]

// ImmutableArray in records - acceptable for API models where collections are small
[assembly: SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "ImmutableArray acceptable for small API response collections", Scope = "type", Target = "~T:ARMEmulator.Models.LoadProgramResponse")]

// HttpClient takes ownership of HttpContent and disposes it
[assembly: SuppressMessage("Reliability", "CA2000:Dispose objects before losing scope", Justification = "HttpClient methods take ownership and dispose content", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Services")]

// File-scoped wrapper records don't need value semantics - they're temporary and converted immediately
[assembly: SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "File-scoped internal wrappers, immediately converted", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Services")]

// AOT/trimming warnings - not using trimming yet, can be addressed when needed
[assembly: SuppressMessage("Trimming", "IL2026:Members annotated with 'RequiresUnreferencedCodeAttribute' require dynamic access", Justification = "Not using trimming in current build", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Services")]
