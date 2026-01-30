using System.Diagnostics.CodeAnalysis;

// Allow underscores in test method names (standard xUnit convention)
[assembly: SuppressMessage("Naming", "CA1707:Identifiers should not contain underscores", Justification = "Test method naming convention", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Tests")]

// Allow array creation in test assertions
[assembly: SuppressMessage("Performance", "CA1861:Avoid constant arrays as arguments", Justification = "Test code readability", Scope = "namespaceanddescendants", Target = "~N:ARMEmulator.Tests")]
