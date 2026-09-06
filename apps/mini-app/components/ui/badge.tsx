import { Chip } from "@heroui/react";

export type BadgeVariant =
  | "neutral"
  | "info"
  | "success"
  | "warning"
  | "error"
  | "yellow"
  | "purple"
  | "blue"
  | "pink"
  | "green"
  | "red";

function chipColor(variant: BadgeVariant): "accent" | "danger" | "success" | "warning" | "default" {
  switch (variant) {
    case "yellow":
    case "warning":
      return "warning";
    case "purple":
      return "accent";
    case "blue":
    case "info":
      return "accent";
    case "pink":
    case "red":
    case "error":
      return "danger";
    case "green":
    case "success":
      return "success";
    default:
      return "default";
  }
}

export function Badge({ variant = "neutral", label }: { variant?: BadgeVariant; label: string }) {
  return (
    <Chip color={chipColor(variant)} size="sm" variant="soft">
      {label}
    </Chip>
  );
}
