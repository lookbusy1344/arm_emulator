using ARMEmulator.Models;
using FluentAssertions;

namespace ARMEmulator.Tests.Models;

public sealed class RegisterStateTests
{
    [Fact]
    public void Create_WithDefaultParameters_ReturnsZeroedRegisters()
    {
        var state = RegisterState.Create();

        state.R0.Should().Be(0u);
        state.R12.Should().Be(0u);
        state.PC.Should().Be(0u);
        state.SP.Should().Be(0u);
        state.LR.Should().Be(0u);
        state.Registers.Should().HaveCount(16);
        state.Registers.Should().AllSatisfy(r => r.Should().Be(0u));
    }

    [Fact]
    public void Create_WithSpecificValues_SetsCorrectRegisters()
    {
        var state = RegisterState.Create(
            r0: 0x100,
            r1: 0x200,
            sp: 0x50000,
            lr: 0x8004,
            pc: 0x8000
        );

        state.R0.Should().Be(0x100u);
        state.R1.Should().Be(0x200u);
        state.SP.Should().Be(0x50000u);
        state.LR.Should().Be(0x8004u);
        state.PC.Should().Be(0x8000u);
    }

    [Fact]
    public void Diff_WithNoChanges_ReturnsEmptySet()
    {
        var state1 = RegisterState.Create(r0: 42, r1: 100);
        var state2 = RegisterState.Create(r0: 42, r1: 100);

        var diff = state1.Diff(state2);

        diff.Should().BeEmpty();
    }

    [Fact]
    public void Diff_WithSingleRegisterChange_ReturnsThatRegister()
    {
        var before = RegisterState.Create(r0: 0, r1: 100);
        var after = RegisterState.Create(r0: 42, r1: 100);

        var diff = after.Diff(before);

        diff.Should().BeEquivalentTo(new[] { "R0" });
    }

    [Fact]
    public void Diff_WithMultipleChanges_ReturnsAllChangedRegisters()
    {
        var before = RegisterState.Create(r0: 0, r1: 100, r2: 200, sp: 0x50000);
        var after = RegisterState.Create(r0: 42, r1: 101, r2: 200, sp: 0x50000);

        var diff = after.Diff(before);

        diff.Should().BeEquivalentTo(new[] { "R0", "R1" });
    }

    [Fact]
    public void Diff_WithCPSRChange_IncludesCPSR()
    {
        var before = RegisterState.Create(cpsr: new CPSRFlags(N: false, Z: false, C: false, V: false));
        var after = RegisterState.Create(cpsr: new CPSRFlags(N: true, Z: false, C: false, V: false));

        var diff = after.Diff(before);

        diff.Should().Contain("CPSR");
    }

    [Fact]
    public void Diff_WithSpecialRegisters_UsesCorrectNames()
    {
        var before = RegisterState.Create(sp: 0x50000, lr: 0, pc: 0x8000);
        var after = RegisterState.Create(sp: 0x4FF00, lr: 0x8004, pc: 0x8004);

        var diff = after.Diff(before);

        diff.Should().BeEquivalentTo(new[] { "SP", "LR", "PC" });
    }

    [Fact]
    public void Indexer_WithValidRegisterName_ReturnsCorrectValue()
    {
        var state = RegisterState.Create(r0: 42, r5: 100, sp: 0x50000);

        state["R0"].Should().Be(42u);
        state["r0"].Should().Be(42u); // Case insensitive
        state["R5"].Should().Be(100u);
        state["SP"].Should().Be(0x50000u);
        state["R13"].Should().Be(0x50000u); // SP alias
    }

    [Fact]
    public void Indexer_WithLR_ReturnsR14Value()
    {
        var state = RegisterState.Create(lr: 0x8004);

        state["LR"].Should().Be(0x8004u);
        state["R14"].Should().Be(0x8004u);
    }

    [Fact]
    public void Indexer_WithPC_ReturnsR15Value()
    {
        var state = RegisterState.Create(pc: 0x8000);

        state["PC"].Should().Be(0x8000u);
        state["R15"].Should().Be(0x8000u);
    }

    [Fact]
    public void Indexer_WithInvalidName_ThrowsArgumentException()
    {
        var state = RegisterState.Create();

        var act = () => state["R16"];

        act.Should().Throw<ArgumentException>()
            .WithMessage("*Unknown register*");
    }

    [Fact]
    public void Indexer_WithInvalidName_InvalidRegister_ThrowsArgumentException()
    {
        var state = RegisterState.Create();

        var act = () => state["FOO"];

        act.Should().Throw<ArgumentException>()
            .WithMessage("*Unknown register*");
    }

    [Fact]
    public void NamedAccessors_ReturnCorrectRegisters()
    {
        var state = RegisterState.Create(
            r0: 1, r1: 2, r2: 3, r3: 4,
            r4: 5, r5: 6, r6: 7, r7: 8,
            r8: 9, r9: 10, r10: 11, r11: 12,
            r12: 13, sp: 14, lr: 15, pc: 16
        );

        state.R0.Should().Be(1u);
        state.R1.Should().Be(2u);
        state.R2.Should().Be(3u);
        state.R3.Should().Be(4u);
        state.R4.Should().Be(5u);
        state.R5.Should().Be(6u);
        state.R6.Should().Be(7u);
        state.R7.Should().Be(8u);
        state.R8.Should().Be(9u);
        state.R9.Should().Be(10u);
        state.R10.Should().Be(11u);
        state.R11.Should().Be(12u);
        state.R12.Should().Be(13u);
        state.SP.Should().Be(14u);
        state.LR.Should().Be(15u);
        state.PC.Should().Be(16u);
    }

    [Fact]
    public void CPSRFlags_DisplayString_FormatsCorrectly()
    {
        var allSet = new CPSRFlags(N: true, Z: true, C: true, V: true);
        var allClear = new CPSRFlags(N: false, Z: false, C: false, V: false);
        var mixed = new CPSRFlags(N: true, Z: false, C: true, V: false);

        allSet.DisplayString.Should().Be("NZCV");
        allClear.DisplayString.Should().Be("----");
        mixed.DisplayString.Should().Be("N-C-");
    }

    [Fact]
    public void RegisterState_IsImmutable()
    {
        var state = RegisterState.Create(r0: 42);
        var registers = state.Registers;

        // ImmutableArray prevents mutation
        registers.Should().BeOfType<ImmutableArray<uint>>();
        state.R0.Should().Be(42u);
    }

    [Fact]
    public void WithExpression_CreatesNewInstance()
    {
        var original = RegisterState.Create(r0: 42);
        var modified = original with { CPSR = new CPSRFlags(N: true, Z: false, C: false, V: false) };

        original.CPSR.N.Should().BeFalse();
        modified.CPSR.N.Should().BeTrue();
        modified.R0.Should().Be(42u);
    }
}
