import React from "react";

interface CardProps {
    children: React.ReactNode;
    className?: string;
    header?: React.ReactNode;
    footer?: React.ReactNode;
}

export function Card({ children, className = "", header, footer }: CardProps) {
    return (
        <div className={`bg-white rounded-2xl shadow-xl overflow-hidden ${className}`}>
            {header && (
                <div className="bg-indigo-600 text-white p-4 text-center font-medium">
                    {header}
                </div>
            )}
            <div className="p-6 md:p-8">
                {children}
            </div>
            {footer && (
                <div className="bg-gray-50 p-4 border-t">
                    {footer}
                </div>
            )}
        </div>
    );
}
