interface DurationToggleProps {
  duration: "24h" | "7d";
  onChange: (duration: "24h" | "7d") => void;
}

export function DurationToggle({ duration, onChange }: DurationToggleProps) {
  return (
    <div className="flex gap-1 border border-gray-200 rounded p-1">
      <button
        onClick={() => onChange("24h")}
        className={`px-3 py-1 text-sm rounded transition-colors ${
          duration === "24h"
            ? "bg-black text-white"
            : "text-gray-600 hover:bg-gray-50"
        }`}
      >
        24h
      </button>
      <button
        onClick={() => onChange("7d")}
        className={`px-3 py-1 text-sm rounded transition-colors ${
          duration === "7d"
            ? "bg-black text-white"
            : "text-gray-600 hover:bg-gray-50"
        }`}
      >
        7d
      </button>
    </div>
  );
}
