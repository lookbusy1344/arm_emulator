using ARMEmulator.Models;
using FluentAssertions;

namespace ARMEmulator.Tests.Models;

public sealed class WatchpointTests
{
	[Fact]
	public void Constructor_StoresAllProperties()
	{
		var watchpoint = new Watchpoint(1, 0x10000, WatchpointType.ReadWrite);

		_ = watchpoint.Id.Should().Be(1);
		_ = watchpoint.Address.Should().Be(0x10000u);
		_ = watchpoint.Type.Should().Be(WatchpointType.ReadWrite);
	}

	[Fact]
	public void Equality_WithSameValues_AreEqual()
	{
		var wp1 = new Watchpoint(1, 0x10000, WatchpointType.Write);
		var wp2 = new Watchpoint(1, 0x10000, WatchpointType.Write);

		_ = wp1.Should().Be(wp2);
	}

	[Fact]
	public void Equality_WithDifferentId_AreNotEqual()
	{
		var wp1 = new Watchpoint(1, 0x10000, WatchpointType.Write);
		var wp2 = new Watchpoint(2, 0x10000, WatchpointType.Write);

		_ = wp1.Should().NotBe(wp2);
	}

	[Theory]
	[InlineData(WatchpointType.Read)]
	[InlineData(WatchpointType.Write)]
	[InlineData(WatchpointType.ReadWrite)]
	public void WatchpointType_HasAllValues(WatchpointType type)
	{
		// Just verify enum values exist
		_ = type.Should().BeOneOf(WatchpointType.Read, WatchpointType.Write, WatchpointType.ReadWrite);
	}
}
