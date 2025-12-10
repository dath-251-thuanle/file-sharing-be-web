import { ReactNode } from "react";
import { AdminTokenBar } from "./AdminTokenBar";

type AdminShellProps = {
  title: string;
  description?: string;
  actions?: ReactNode;
  children: ReactNode;
};

export function AdminShell({ title, description, actions, children }: AdminShellProps) {
  return (
    <div className="mx-auto max-w-3xl p-6">
      <div className="mb-6 flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">{title}</h1>
          {description && (
            <p className="text-sm text-gray-500">{description}</p>
          )}
        </div>
        {actions}
      </div>

      <AdminTokenBar />

      {children}
    </div>
  );
}
