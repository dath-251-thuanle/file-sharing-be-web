import React from "react";

interface AlertProps {
    type: "success" | "error" | "warning" | "info";
    message: string;
    className?: string;
}

export function Alert({ type, message, className = "" }: AlertProps) {
    const styles = {
        success: "bg-green-100 border-green-300 text-green-800",
        error: "bg-red-100 border-red-300 text-red-800",
        warning: "bg-yellow-100 border-yellow-300 text-yellow-800",
        info: "bg-blue-100 border-blue-300 text-blue-800",
    };

    const icons = {
        success: "✅",
        error: "❌",
        warning: "⚠️",
        info: "ℹ️",
    };

    return (
        <div className={`p-4 border rounded-lg ${styles[type]} ${className}`}>
            <span className="mr-2">{icons[type]}</span>
            {message}
        </div>
    );
}
