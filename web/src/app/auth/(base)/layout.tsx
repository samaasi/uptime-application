import * as React from "react";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-black text-white relative overflow-hidden">
      <div className="absolute inset-0 w-full h-full bg-cover bg-center bg-no-repeat"
        style={{
          backgroundImage: "url('/auth-background.svg')",
        }}
      />
      
      {/* Content overlay */}
      <div className="relative z-10 flex flex-col items-center justify-center min-h-screen">
        <h2 className="text-4xl font-bold mb-8">Uptime</h2>
        <div className="">
          { children }
        </div>
        <p className="text-center text-xs text-neutral-700 mt-8 uppercase">© 2025 Uptime.</p>
      </div>
    </div>
  );
}