"use client";

import {
  createElement,
  isValidElement,
  useCallback,
  useMemo,
  useRef,
  type ReactElement,
  type ReactNode,
  type Ref,
} from "react";
import {
  ActionIcon,
  playActionIcon,
  type ActionIconHandle,
  type ActionIconName,
} from "@/components/action-icon";

function mergeRefs<T>(...refs: Array<Ref<T> | undefined | null>) {
  return (value: T | null) => {
    for (const ref of refs) {
      if (!ref) continue;
      if (typeof ref === "function") ref(value);
      else ref.current = value;
    }
  };
}

type ActionIconElement = ReactElement<{
  name: ActionIconName;
  size?: number;
  className?: string;
  ref?: Ref<ActionIconHandle>;
}>;

function isActionIconElement(icon: ReactNode): icon is ActionIconElement {
  if (!isValidElement(icon)) return false;
  const type = icon.type as { displayName?: string };
  return icon.type === ActionIcon || type.displayName === "ActionIcon";
}

/** Wire ActionIcon ref + parent press handler so tap anywhere on the button animates the icon. */
export function useActionIconTrigger(icon: ReactNode | undefined) {
  const iconRef = useRef<ActionIconHandle | null>(null);

  const onParentPointerDown = useCallback(() => {
    playActionIcon(iconRef.current);
  }, []);

  const wiredIcon = useMemo(() => {
    if (!isActionIconElement(icon)) return icon;

    const { name, size, className, ref: existingRef } = icon.props;
    return createElement(ActionIcon, {
      ref: mergeRefs(existingRef, iconRef),
      name,
      size,
      className,
    });
  }, [icon]);

  return {
    icon: wiredIcon,
    onParentPointerDown,
    hasActionIcon: isActionIconElement(icon),
  };
}

/** Stable hook for nav/button icons: returns icon node + press handler. */
export function usePressableActionIcon(name: ActionIconName, size = 18) {
  const iconRef = useRef<ActionIconHandle | null>(null);
  const icon = useMemo(
    () => createElement(ActionIcon, { ref: iconRef, name, size }),
    [name, size]
  );
  const onPress = useCallback(() => {
    playActionIcon(iconRef.current);
  }, []);

  return { icon, onPress };
}
