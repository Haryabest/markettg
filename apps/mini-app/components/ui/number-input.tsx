"use client";

import { NumberField } from "@heroui/react";
import { Minus, Plus } from "lucide-react";

export function NumberInput({
  label,
  value,
  onChange,
  min = 1,
  max = 99,
  width,
  isLabelHidden,
}: {
  label?: string;
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  hasNumberSteppers?: boolean;
  width?: number | string;
  isLabelHidden?: boolean;
  isIntegerOnly?: boolean;
}) {
  return (
    <NumberField
      aria-label={isLabelHidden ? label || "Количество" : undefined}
      value={value}
      onChange={(v) => onChange(Number(v))}
      minValue={min}
      maxValue={max}
      className={typeof width === "number" ? `w-[${width}px]` : undefined}
      style={typeof width === "number" ? { width } : undefined}
    >
      <NumberField.Group>
        <NumberField.DecrementButton>
          <Minus size={14} />
        </NumberField.DecrementButton>
        <NumberField.Input />
        <NumberField.IncrementButton>
          <Plus size={14} />
        </NumberField.IncrementButton>
      </NumberField.Group>
    </NumberField>
  );
}
