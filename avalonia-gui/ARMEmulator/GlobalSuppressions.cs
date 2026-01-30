using System.Diagnostics.CodeAnalysis;

// Domain exceptions use primary constructors with specific parameters - don't need standard exception constructors
[assembly: SuppressMessage("Design", "RCS1194:Implement exception constructors", Justification = "Domain exceptions use primary constructors", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Services")]

// ImmutableArray in records - acceptable for API models where collections are small
[assembly: SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "ImmutableArray acceptable for small API response collections", Scope = "type", Target = "~T:ARMEmulator.Models.LoadProgramResponse")]
