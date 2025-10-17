import Link from "next/link";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function SignUpPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-black text-white">
      <div className="grid md:grid-cols-2 w-full max-w-5xl">
        <Card className="bg-neutral-900 border-neutral-800 shadow-lg p-8 rounded-2xl">
          <CardHeader>
            <CardTitle className="text-2xl font-semibold">Create an account</CardTitle>
            <p className="text-sm text-neutral-400">
              Already have an account?{" "}
              <Link href="/sign-in" className="underline hover:text-white">
                Sign in
              </Link>
            </p>
          </CardHeader>
          <CardContent>
            <form className="space-y-6">
              <div className="space-y-2">
                <Label htmlFor="name">Full name</Label>
                <Input
                  id="name"
                  type="text"
                  placeholder="Enter full name"
                  className="bg-neutral-800 border-neutral-700 text-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="Enter email"
                  className="bg-neutral-800 border-neutral-700 text-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="Enter password"
                  className="bg-neutral-800 border-neutral-700 text-white"
                />
              </div>
              <Button className="w-full bg-blue-600 hover:bg-blue-700">
                Create account
              </Button>
            </form>
            <p className="mt-6 text-xs text-neutral-500">
              By signing up, you agree to our{" "}
              <Link href="#" className="underline hover:text-white">
                Terms of Service
              </Link>
              .
            </p>
          </CardContent>
        </Card>

        {/* Right side illustration / ID card */}
        <div className="hidden md:flex items-center justify-center bg-neutral-950">
          <div className="flex flex-col items-center border border-neutral-800 rounded-2xl p-8 text-neutral-400">
            <div className="w-24 h-24 border border-dashed border-neutral-700 flex items-center justify-center rounded-md mb-4">
              <span className="text-xs">+<br />Avatar<br />Max 2MB</span>
            </div>
            <p className="text-lg font-medium">New User</p>
            <div className="mt-4 text-[10px] text-green-500 font-mono">
              0199ed92-5f0b-7990-9bcd-9704b2d40b95
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
