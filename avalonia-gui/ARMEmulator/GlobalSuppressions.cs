using System.Diagnostics.CodeAnalysis;

// Domain exceptions use primary constructors with specific parameters - standard exception constructors not needed
[assembly: SuppressMessage("Design", "RCS1194:Implement exception constructors", Justification = "Domain exceptions use primary constructors with domain-specific parameters", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Services")]

// ImmutableArray in records - acceptable for API models where collections are small
// Note: ImmutableArray<T> is a struct wrapping a reference, so doesn't have true value semantics,
// but for small API response collections this is acceptable and more performant than IReadOnlyList<T>
[assembly: SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "ImmutableArray acceptable for small API response collections", Scope = "type", Target = "~T:ARMEmulator.Models.LoadProgramResponse")]

// HttpClient takes ownership of HttpContent and disposes it
[assembly: SuppressMessage("Reliability", "CA2000:Dispose objects before losing scope", Justification = "HttpClient methods take ownership and dispose content", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Services")]

// Internal wrapper records use arrays for JSON deserialization and are immediately converted to ImmutableArray
// These are never exposed outside the Services namespace
[assembly: SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "Internal wrappers for JSON deserialization, immediately converted to ImmutableArray", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Services")]
