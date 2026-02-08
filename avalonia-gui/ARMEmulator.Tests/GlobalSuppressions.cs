using System.Diagnostics.CodeAnalysis;

// Allow underscores in test method names (standard xUnit convention)
[assembly: SuppressMessage("Naming", "CA1707:Identifiers should not contain underscores", Justification = "Test method naming convention", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Tests")]

// Allow array creation in test assertions
[assembly: SuppressMessage("Performance", "CA1861:Avoid constant arrays as arguments", Justification = "Test code readability", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Tests")]

// Trimming warnings not relevant for tests
[assembly: SuppressMessage("Trimming", "IL2026:Members annotated with 'RequiresUnreferencedCodeAttribute' require dynamic access", Justification = "Not using trimming in tests", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Tests")]

// Expression values don't need to be used in tests (FluentAssertions returns values for chaining)
[assembly: SuppressMessage("Style", "IDE0058:Expression value is never used", Justification = "Test assertions return values for chaining but we don't always chain them", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Tests")]

// Single-line if statements are acceptable in tests for brevity
[assembly: SuppressMessage("Style", "IDE0011:Add braces", Justification = "Test code readability", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Tests")]
