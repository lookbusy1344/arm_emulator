namespace ARMEmulator.Tests;

using System.CodeDom.Compiler;
using System.Collections.Frozen;
using System.Reflection;
using System.Reflection.Metadata;
using System.Reflection.Metadata.Ecma335;
using System.Reflection.PortableExecutable;
using System.Runtime.CompilerServices;
using AwesomeAssertions;
using CodeStyle;

/// <summary>
///     Architecture guard for the .NET performance-guideline size ceiling on value types (Framework
///     Design Guidelines: prefer a class once a struct's instance size passes roughly 16-24 bytes,
///     because every pass-by-value copies the whole thing). Uses reflection over the compiled
///     production assemblies rather than Roslyn: instance layout — decimal is 16 bytes, an
///     <c>Instant</c> is 16, a reference field is always 8 — is a runtime fact a syntax tree cannot
///     see. Derives the production assembly set from every project under <c>src</c>, then scans every
///     non-generic struct/record struct in those assemblies, plus every closed generic struct
///     instantiation encoded in the compiled assemblies' ECMA-335
///     <c>TypeSpec</c> metadata. This measures the actual runtime layout of, for example,
///     <c>Wrapper&lt;decimal&gt;</c>, rather than pretending the open <c>Wrapper&lt;T&gt;</c> definition has
///     the same layout as <c>Wrapper&lt;object&gt;</c>.
///     <para>
///         Generated structs (<c>[GeneratedCode]</c>/<c>[CompilerGenerated]</c> — e.g. a
///         <c>[LoggerMessage]</c> parameter carrier or an async state machine) are not authored types
///         an engineer can redesign. A struct authored on purpose above the ceiling carries
///         <see cref="LargeStructAttribute" /> (<c>ARMEmulator.CodeStyle</c>) with its own
///         reviewed reason — attach only with explicit user permission, per that attribute's own
///         summary.
///     </para>
///     Razor-compiled types are classes, not structs, so <c>.cshtml</c> needs no separate coverage.
/// </summary>
public sealed class CodeStyle_StructSize
{
	[Fact]
	public void Repository_value_types_do_not_exceed_the_24_byte_size_guideline()
	{
		var violations = ValueTypeSizeGuard.FindViolations(ValueTypeSizeGuard.ProductionAssemblyNames).ToArray();

		violations.Should().BeEmpty(
			"a value type over {0} bytes is copied wholesale on every pass-by-value -- shrink the fields, wrap " +
			"the bulk behind a reference, or convert to a class:{1}{2}",
			ValueTypeSizeGuard.MaxSizeInBytes,
			Environment.NewLine,
			string.Join(Environment.NewLine, violations));
	}

	[Fact]
	public void Struct_at_or_under_the_ceiling_is_not_a_violation() =>
		ValueTypeSizeGuard.Measure(typeof(EightByteStruct)).Should().Be(8);

	[Fact]
	public void Struct_over_the_ceiling_is_a_violation()
	{
		var size = ValueTypeSizeGuard.Measure(typeof(ThirtyTwoByteStruct));

		size.Should().BeGreaterThan(ValueTypeSizeGuard.MaxSizeInBytes);
	}

	[Fact]
	public void Oversized_record_struct_is_measured_the_same_as_an_oversized_struct()
	{
		var size = ValueTypeSizeGuard.Measure(typeof(ThirtyTwoByteRecordStruct));

		size.Should().BeGreaterThan(ValueTypeSizeGuard.MaxSizeInBytes);
	}

	[Fact]
	public void Enum_is_not_treated_as_a_value_type_to_measure() =>
		ValueTypeSizeGuard.ConcreteValueTypes([typeof(ExampleEnum).Assembly.GetName().Name!])
						  .Should().NotContain(typeof(ExampleEnum));

	[Fact]
	public void Open_generic_struct_cannot_be_measured_as_a_concrete_runtime_layout()
	{
		var measure = () => ValueTypeSizeGuard.Measure(typeof(GenericReferenceWrapper<>));

		measure.Should().Throw<ArgumentException>().WithParameterName("type");
	}

	[Fact]
	public void Open_generic_definition_is_not_treated_as_a_concrete_value_type()
	{
		var assemblyName = typeof(CodeStyle_StructSize).Assembly.GetName().Name!;

		ValueTypeSizeGuard.ConcreteValueTypes([assemblyName]).Should().NotContain(typeof(GenericInlineWrapper<>));
	}

	[Fact]
	public void Closed_generic_struct_instantiation_is_measured_using_its_actual_type_arguments()
	{
		_ = ClosedGenericUsage();
		var assemblyName = typeof(CodeStyle_StructSize).Assembly.GetName().Name!;

		ValueTypeSizeGuard.FindViolations([assemblyName])
						  .Should().Contain(violation =>
							  violation.Contains(nameof(GenericInlineWrapper<ThirtyTwoByteStruct>), StringComparison.Ordinal)
							  && violation.Contains("32 bytes", StringComparison.Ordinal));
	}

	[Fact]
	public void Oversized_ref_struct_requires_a_reviewed_exception()
	{
		var assemblyName = typeof(CodeStyle_StructSize).Assembly.GetName().Name!;

		ValueTypeSizeGuard.FindViolations([assemblyName])
						  .Should().Contain(violation => violation.Contains(nameof(LargeRefStruct), StringComparison.Ordinal));
	}

	[Fact]
	public void Every_production_assembly_is_in_the_size_scan() =>
		ValueTypeSizeGuard.ProductionAssemblyNames.Should().Contain(["ARMEmulator"]);

	[Fact]
	public void Struct_carrying_LargeStructAttribute_is_excluded_from_the_scan()
	{
		var assemblyName = typeof(CodeStyle_StructSize).Assembly.GetName().Name!;

		ValueTypeSizeGuard.FindViolations([assemblyName])
						  .Should().NotContain(violation => violation.Contains(nameof(ReviewedLargeStruct), StringComparison.Ordinal));
	}

	[Fact]
	public void Source_generator_emitted_struct_is_excluded_from_the_scan()
	{
		// Mirrors the shape of a [LoggerMessage] source-generated parameter-carrier struct: the
		// generator, not an author, decides its field list and size.
		var assemblyName = typeof(CodeStyle_StructSize).Assembly.GetName().Name!;

		ValueTypeSizeGuard.ConcreteValueTypes([assemblyName]).Should().NotContain(typeof(SourceGeneratedLargeStruct));
	}

	[Fact]
	public void Compiler_generated_state_machine_struct_is_excluded_from_the_scan()
	{
		var assemblyName = typeof(CodeStyle_StructSize).Assembly.GetName().Name!;
		var stateMachineTypes = ValueTypeSizeGuard.ConcreteValueTypes([assemblyName])
												  .Where(static type => type.Name.Contains("AsyncMethodWithManyLocalsAsync", StringComparison.Ordinal));

		// The compiler generates a struct holding every local below (well past the 24-byte ceiling) to
		// drive the state machine. It is an artifact of how `async` compiles, not an authored type, so
		// the [CompilerGenerated] filter must keep it out of the scan entirely -- reporting it would be
		// a false positive no author can fix by hand.
		stateMachineTypes.Should().BeEmpty();
	}

	private static async Task AsyncMethodWithManyLocalsAsync()
	{
		await Task.Yield();
		var a = 1L;
		var b = 2L;
		var c = 3L;
		var d = 4L;
		var e = 5L;
		_ = a + b + c + d + e;
	}

	private static GenericInlineWrapper<ThirtyTwoByteStruct> ClosedGenericUsage() => default;

	private struct EightByteStruct
	{
		public long Value;

		public EightByteStruct(long value) => Value = value;
	}

	private struct ThirtyTwoByteStruct
	{
		public long A;
		public long B;
		public long C;
		public long D;

		public ThirtyTwoByteStruct(long a, long b, long c, long d)
		{
			A = a;
			B = b;
			C = c;
			D = d;
		}
	}

	private record struct ThirtyTwoByteRecordStruct(long A, long B, long C, long D);

	private enum ExampleEnum
	{
		None,
	}

	private readonly struct GenericReferenceWrapper<T>(T[]? items)
	{
		private readonly T[]? items = items;

		public bool IsEmpty => items is null or [];
	}

	private readonly struct GenericInlineWrapper<T>(T value)
	{
		public T Value { get; } = value;
	}

	private readonly ref struct LargeRefStruct
	{
		private readonly long a;
		private readonly long b;
		private readonly long c;
		private readonly long d;

		public LargeRefStruct(long a, long b, long c, long d)
		{
			this.a = a;
			this.b = b;
			this.c = c;
			this.d = d;
		}

		public long Sum => a + b + c + d;
	}

	[LargeStruct("Fixture proving FindViolations respects a reviewed exception.")]
	private readonly struct ReviewedLargeStruct
	{
		private readonly long a;
		private readonly long b;
		private readonly long c;
		private readonly long d;

		public ReviewedLargeStruct(long a, long b, long c, long d)
		{
			this.a = a;
			this.b = b;
			this.c = c;
			this.d = d;
		}

		public long Sum => a + b + c + d;
	}

	[GeneratedCode("JobTrack.ArchitectureTests.CodeStyle_StructSize", "1.0")]
	private readonly struct SourceGeneratedLargeStruct
	{
		private readonly long a;
		private readonly long b;
		private readonly long c;
		private readonly long d;

		public SourceGeneratedLargeStruct(long a, long b, long c, long d)
		{
			this.a = a;
			this.b = b;
			this.c = c;
			this.d = d;
		}

		public long Sum => a + b + c + d;
	}
}

internal static class ValueTypeSizeGuard
{
	public const int MaxSizeInBytes = 24;

	public static readonly FrozenSet<string> ProductionAssemblyNames = Directory
																	   .EnumerateFiles(Path.Combine(Harness.RepoRoot, "ARMEmulator"), "*.csproj", SearchOption.AllDirectories)
																	   .Select(static path => Path.GetFileNameWithoutExtension(path)!)
																	   .ToFrozenSet(StringComparer.Ordinal);

	private static readonly MethodInfo SizeOfMethod =
		typeof(Unsafe).GetMethod(nameof(Unsafe.SizeOf), BindingFlags.Public | BindingFlags.Static)!;

	public static IEnumerable<string> FindViolations(IEnumerable<string> assemblyNames) =>
		ConcreteValueTypes(assemblyNames)
			.Where(static type => !Attribute.IsDefined(type, typeof(LargeStructAttribute)))
			.Select(type => (Type: type, Size: Measure(type)))
			.Where(measurement => measurement.Size > MaxSizeInBytes)
			.Select(measurement => Describe(measurement.Type, measurement.Size))
			.Order(StringComparer.Ordinal);

	public static IEnumerable<Type> ConcreteValueTypes(IEnumerable<string> assemblyNames)
	{
		var includedAssemblyNames = assemblyNames.ToFrozenSet(StringComparer.Ordinal);
		var assemblies = includedAssemblyNames.Select(Assembly.Load).ToArray();

		return assemblies
			   .SelectMany(assembly => assembly.GetTypes()
											   .Where(static type => !type.IsGenericTypeDefinition)
											   .Concat(ClosedGenericTypes(assembly, includedAssemblyNames)))
			   .Distinct()
			   .Where(static type => type.IsValueType && !type.IsEnum)
			   .Where(static type => !Attribute.IsDefined(type, typeof(CompilerGeneratedAttribute)))
			   .Where(static type => !Attribute.IsDefined(type, typeof(GeneratedCodeAttribute)));
	}

	private static Type[] ClosedGenericTypes(Assembly assembly, FrozenSet<string> includedAssemblyNames)
	{
		using var stream = File.OpenRead(assembly.Location);
		using var peReader = new PEReader(stream);
		var metadataReader = peReader.GetMetadataReader();
		var typeSpecificationCount = metadataReader.GetTableRowCount(TableIndex.TypeSpec);

		return Enumerable.Range(1, typeSpecificationCount)
						 .Select(MetadataTokens.TypeSpecificationHandle)
						 .Select(handle => ResolveClosedTypeSpecification(assembly.ManifestModule, MetadataTokens.GetToken(handle)))
						 .Where(static type => type is not null)
						 .Cast<Type>()
						 .SelectMany(ContainedTypes)
						 .Where(static type => type.IsConstructedGenericType && !type.ContainsGenericParameters)
						 .Where(type => includedAssemblyNames.Contains(type.GetGenericTypeDefinition().Assembly.GetName().Name!))
						 .ToArray();
	}

	private static Type? ResolveClosedTypeSpecification(Module module, int metadataToken)
	{
		try {
			return module.ResolveType(metadataToken);
		}
		catch (ArgumentException) {
			// A TypeSpec containing a type-level or method-level generic parameter has no single
			// runtime layout until its declaring generic context is itself closed.
			return null;
		}
	}

	private static IEnumerable<Type> ContainedTypes(Type type)
	{
		yield return type;
		foreach (var argument in type.GetGenericArguments()) {
			foreach (var contained in ContainedTypes(argument)) {
				yield return contained;
			}
		}
	}

	/// <summary>
	///     The type's runtime instance size, in bytes, via <see cref="Unsafe.SizeOf{T}" /> — unlike
	///     <c>Marshal.SizeOf</c>, this is the true managed layout size (a <see langword="bool" /> is 1
	///     byte, not the 4-byte marshaled default) and, unlike C#'s <c>sizeof</c> operator, it carries no
	///     <c>unmanaged</c> constraint, so a struct holding a reference-typed field measures correctly too.
	///     Rejects an open generic because it has no single concrete runtime layout;
	///     <see cref="ConcreteValueTypes" /> finds the closed instantiations recorded by the compiled code.
	/// </summary>
	public static int Measure(Type type)
	{
		if (type.ContainsGenericParameters) {
			throw new ArgumentException("An open generic type has no concrete runtime layout.", nameof(type));
		}

		// Justification: this repo enables trim analysis; the test assembly is never trimmed
#pragma warning disable IL2060
		return (int)SizeOfMethod.MakeGenericMethod(type).Invoke(null, null)!;
#pragma warning restore IL2060
	}

	private static string Describe(Type type, int size) =>
		$"{type.FullName ?? type.Name}: {size} bytes (> {MaxSizeInBytes})";
}
