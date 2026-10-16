import { ref } from "vue";

export type ToolScope = "workspace" | "user";

const activeScope = ref<ToolScope>("workspace");

export function useToolScope() {
  function setScope(scope: ToolScope) {
    activeScope.value = scope;
  }

  return {
    activeScope,
    setScope,
  };
}
