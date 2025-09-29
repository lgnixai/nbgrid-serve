import { useMemo } from "react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export interface SpaceOption {
  id: string;
  name: string;
  bases: Array<{ id: string; name: string; tables: string[] }>;
}

interface SpaceBaseSelectorProps {
  spaces: SpaceOption[];
  spaceId?: string;
  baseId?: string;
  onChange: (spaceId: string, baseId: string) => void;
}

export const SpaceBaseSelector = ({ spaces, spaceId, baseId, onChange }: SpaceBaseSelectorProps) => {
  const currentSpace = useMemo(() => spaces.find(s => s.id === spaceId) ?? spaces[0], [spaces, spaceId]);
  const bases = currentSpace?.bases ?? [];

  return (
    <div className="flex items-center gap-2">
      <Select
        value={spaceId ?? currentSpace?.id}
        onValueChange={(newSpaceId) => {
          const s = spaces.find(sp => sp.id === newSpaceId) ?? spaces[0];
          const firstBase = s?.bases?.[0]?.id;
          onChange(newSpaceId, firstBase ?? "");
        }}
      >
        <SelectTrigger className="w-44 h-8 bg-obsidian-bg border-obsidian-border text-obsidian-text">
          <SelectValue placeholder="选择 Space" />
        </SelectTrigger>
        <SelectContent className="bg-obsidian-surface border-obsidian-border text-obsidian-text">
          {spaces.map(s => (
            <SelectItem key={s.id} value={s.id} className="cursor-pointer">
              {s.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select
        value={baseId ?? bases[0]?.id}
        onValueChange={(newBaseId) => {
          const sid = (spaceId ?? currentSpace?.id) as string;
          onChange(sid, newBaseId);
        }}
      >
        <SelectTrigger className="w-44 h-8 bg-obsidian-bg border-obsidian-border text-obsidian-text">
          <SelectValue placeholder="选择 Base" />
        </SelectTrigger>
        <SelectContent className="bg-obsidian-surface border-obsidian-border text-obsidian-text">
          {bases.map(b => (
            <SelectItem key={b.id} value={b.id} className="cursor-pointer">
              {b.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
};


