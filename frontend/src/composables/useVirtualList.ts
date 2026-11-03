// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed, onMounted, onBeforeUnmount, getCurrentInstance, type Ref } from "vue";

export interface VirtualItem<T> {
  item: T;
  index: number;
  top: number;
  height: number;
}

export interface UseVirtualListOptions<T> {
  items: Ref<T[]>;
  itemHeight: number | ((item: T, index: number) => number);
  containerRef?: Ref<HTMLElement | null>;
  overscan?: number;
}

export function useVirtualList<T>(options: UseVirtualListOptions<T>) {
  const { items, itemHeight, containerRef, overscan = 8 } = options;

  const scrollTop = ref(0);
  const viewportHeight = ref(600);

  const isFixedHeight = typeof itemHeight === "number";
  const fixedHeightVal = isFixedHeight ? (itemHeight as number) : 0;

  // Precomputed offsets for variable heights
  const offsets = computed<number[]>(() => {
    const list = items.value;
    const n = list.length;
    const res = new Array<number>(n + 1);
    res[0] = 0;

    if (isFixedHeight) {
      for (let i = 0; i < n; i++) {
        res[i + 1] = (i + 1) * fixedHeightVal;
      }
    } else {
      const fn = itemHeight as (item: T, index: number) => number;
      let accum = 0;
      for (let i = 0; i < n; i++) {
        accum += fn(list[i], i);
        res[i + 1] = accum;
      }
    }
    return res;
  });

  const totalHeight = computed(() => {
    const off = offsets.value;
    return off.length > 0 ? off[off.length - 1] : 0;
  });

  // Binary search to find the lower bound index where offset >= target
  function findStartIndex(target: number): number {
    const off = offsets.value;
    const n = items.value.length;
    if (n === 0) return 0;
    if (isFixedHeight && fixedHeightVal > 0) {
      return Math.min(n - 1, Math.max(0, Math.floor(target / fixedHeightVal)));
    }

    let low = 0;
    let high = n - 1;
    let ans = 0;

    while (low <= high) {
      const mid = Math.floor((low + high) / 2);
      if (off[mid + 1] > target) {
        ans = mid;
        high = mid - 1;
      } else {
        low = mid + 1;
      }
    }
    return ans;
  }

  // Binary search to find the upper bound index where offset >= target
  function findEndIndex(target: number): number {
    const off = offsets.value;
    const n = items.value.length;
    if (n === 0) return 0;
    if (isFixedHeight && fixedHeightVal > 0) {
      return Math.min(n, Math.ceil(target / fixedHeightVal));
    }

    let low = 0;
    let high = n - 1;
    let ans = n;

    while (low <= high) {
      const mid = Math.floor((low + high) / 2);
      if (off[mid] >= target) {
        ans = mid;
        high = mid - 1;
      } else {
        low = mid + 1;
      }
    }
    return ans;
  }

  const range = computed(() => {
    const total = items.value.length;
    if (total === 0) {
      return { start: 0, end: 0 };
    }

    const sTop = scrollTop.value;
    const vHeight = viewportHeight.value;

    const rawStart = findStartIndex(sTop);
    const rawEnd = findEndIndex(sTop + vHeight);

    const start = Math.max(0, rawStart - overscan);
    const end = Math.min(total, rawEnd + overscan);

    return { start, end };
  });

  const visibleItems = computed<VirtualItem<T>[]>(() => {
    const { start, end } = range.value;
    const list = items.value;
    const off = offsets.value;
    const result: VirtualItem<T>[] = [];

    for (let i = start; i < end; i++) {
      const top = off[i];
      const height = off[i + 1] - off[i];
      result.push({
        item: list[i],
        index: i,
        top,
        height,
      });
    }

    return result;
  });

  let rafId: number | null = null;
  function handleScroll(e?: Event) {
    if (rafId !== null) return;
    rafId = requestAnimationFrame(() => {
      rafId = null;
      if (containerRef?.value) {
        scrollTop.value = containerRef.value.scrollTop;
      } else if (e && e.target) {
        scrollTop.value = (e.target as HTMLElement).scrollTop;
      }
    });
  }

  let resizeObserver: ResizeObserver | null = null;

  function updateContainerSize() {
    if (containerRef?.value) {
      viewportHeight.value = containerRef.value.clientHeight || 600;
      scrollTop.value = containerRef.value.scrollTop;
    }
  }

  if (getCurrentInstance()) {
    onMounted(() => {
      if (containerRef?.value) {
        updateContainerSize();
        containerRef.value.addEventListener("scroll", handleScroll, { passive: true });

        if (typeof ResizeObserver !== "undefined") {
          resizeObserver = new ResizeObserver(() => {
            updateContainerSize();
          });
          resizeObserver.observe(containerRef.value);
        }
      }
    });

    onBeforeUnmount(() => {
      if (rafId !== null) {
        cancelAnimationFrame(rafId);
        rafId = null;
      }
      if (containerRef?.value) {
        containerRef.value.removeEventListener("scroll", handleScroll);
      }
      if (resizeObserver) {
        resizeObserver.disconnect();
        resizeObserver = null;
      }
    });
  }

  function scrollToIndex(index: number, align: "start" | "center" | "end" | "auto" = "auto") {
    const total = items.value.length;
    if (index < 0 || index >= total || !containerRef?.value) return;

    const off = offsets.value;
    const itemTop = off[index];
    const itemHeightVal = off[index + 1] - off[index];
    const currentScrollTop = containerRef.value.scrollTop;
    const vHeight = viewportHeight.value;

    let targetScrollTop = currentScrollTop;

    if (align === "start") {
      targetScrollTop = itemTop;
    } else if (align === "end") {
      targetScrollTop = itemTop + itemHeightVal - vHeight;
    } else if (align === "center") {
      targetScrollTop = itemTop + itemHeightVal / 2 - vHeight / 2;
    } else {
      // auto
      if (itemTop < currentScrollTop) {
        targetScrollTop = itemTop;
      } else if (itemTop + itemHeightVal > currentScrollTop + vHeight) {
        targetScrollTop = itemTop + itemHeightVal - vHeight;
      }
    }

    containerRef.value.scrollTop = Math.max(0, Math.min(totalHeight.value - vHeight, targetScrollTop));
    scrollTop.value = containerRef.value.scrollTop;
  }

  return {
    scrollTop,
    viewportHeight,
    totalHeight,
    range,
    visibleItems,
    handleScroll,
    scrollToIndex,
    updateContainerSize,
  };
}
