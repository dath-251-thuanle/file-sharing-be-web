type NumberFieldProps = {
  label: string;
  description?: string;
  value?: number;
  min?: number;
  onChange: (value: number | undefined) => void;
};

export function NumberField({
  label,
  description,
  value,
  min = 1,
  onChange,
}: NumberFieldProps) {
  return (
    <div className="space-y-1.5">
      <label className="block text-sm font-medium">{label}</label>
      <input
        type="number"
        min={min}
        value={value ?? ""}
        onChange={(e) => {
          const v = e.target.value;
          onChange(v === "" ? undefined : Number(v));
        }}
        className="block w-full rounded-md border px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-blue-500"
      />
      {description && (
        <p className="text-xs text-gray-500">{description}</p>
      )}
    </div>
  );
}
