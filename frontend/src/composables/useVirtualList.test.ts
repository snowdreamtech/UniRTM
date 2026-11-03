// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { ref } from "vue";
import { useVirtualList } from "./useVirtualList";

describe("useVirtualList composable", () => {
  it("calculates total height and visible items correctly for fixed height", () => {
    const items = ref(Array.from({ length: 100 }, (_, i) => `Item ${i}`));
    const { totalHeight, visibleItems, range, scrollTop } = useVirtualList({
      items,
      itemHeight: 30,
      overscan: 2,
    });

    expect(totalHeight.value).toBe(3000);
    // Default viewportHeight is 600, scrollTop is 0
    // 600 / 30 = 20 items. range: start = 0, end = 20 + 2 = 22
    expect(range.value.start).toBe(0);
    expect(range.value.end).toBe(22);
    expect(visibleItems.value.length).toBe(22);
    expect(visibleItems.value[0].item).toBe("Item 0");
    expect(visibleItems.value[0].top).toBe(0);
    expect(visibleItems.value[0].height).toBe(30);

    // Scroll down by 300px (10 items)
    scrollTop.value = 300;
    // start should be 10 - 2 = 8, end should be 30 + 2 = 32
    expect(range.value.start).toBe(8);
    expect(range.value.end).toBe(32);
    expect(visibleItems.value[0].index).toBe(8);
    expect(visibleItems.value[0].top).toBe(240);
  });

  it("handles empty items list gracefully", () => {
    const items = ref<string[]>([]);
    const { totalHeight, visibleItems, range } = useVirtualList({
      items,
      itemHeight: 25,
    });

    expect(totalHeight.value).toBe(0);
    expect(range.value.start).toBe(0);
    expect(range.value.end).toBe(0);
    expect(visibleItems.value.length).toBe(0);
  });

  it("supports dynamic/variable item height function", () => {
    const items = ref([
      { id: 1, h: 20 },
      { id: 2, h: 40 },
      { id: 3, h: 60 },
      { id: 4, h: 80 },
    ]);
    const { totalHeight, visibleItems } = useVirtualList({
      items,
      itemHeight: (item) => item.h,
      overscan: 0,
    });

    expect(totalHeight.value).toBe(20 + 40 + 60 + 80);
    expect(visibleItems.value.length).toBe(4);
    expect(visibleItems.value[0].top).toBe(0);
    expect(visibleItems.value[0].height).toBe(20);
    expect(visibleItems.value[1].top).toBe(20);
    expect(visibleItems.value[1].height).toBe(40);
    expect(visibleItems.value[2].top).toBe(60);
    expect(visibleItems.value[2].height).toBe(60);
  });

  it("supports scrollToIndex navigation", () => {
    const items = ref(Array.from({ length: 50 }, (_, i) => `Item ${i}`));
    const fakeContainer = {
      clientHeight: 300,
      scrollTop: 0,
    } as any;

    const containerRef = ref<HTMLElement | null>(fakeContainer);
    const { scrollToIndex, totalHeight, updateContainerSize } = useVirtualList({
      items,
      itemHeight: 20,
      containerRef,
    });
    updateContainerSize();

    expect(totalHeight.value).toBe(1000);

    scrollToIndex(10, "start");
    expect(fakeContainer.scrollTop).toBe(200);

    scrollToIndex(20, "end");
    // itemTop = 400, itemHeight = 20, vHeight = 300 => 400 + 20 - 300 = 120
    expect(fakeContainer.scrollTop).toBe(120);
  });
});
