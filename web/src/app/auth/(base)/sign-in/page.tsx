import Link from "next/link";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function SignInPage() {
  return (
    <Card className="w-full max-w-md bg-neutral-900 border-neutral-800 shadow-sm">
        <CardHeader>
            <div className="flex justify-center items-center">
            <div className="border-4 border-black rounded-2xl w-8"/>
            </div>
          <CardTitle className="text-white text-center text-2xl font-semibold mt-10">Sign in</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="email" className="font-semibold text-gray-300">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="Enter your email"
                className="bg-neutral-800 border-neutral-700 text-white placeholder:text-gray-600"
              />
            </div>
            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <Label htmlFor="password" className="font-semibold text-gray-300">Password</Label>
                <Link
                  href="#"
                  className="text-sm text-neutral-400 hover:text-white transition"
                >
                  Forgot password?
                </Link>
              </div>
              <Input
                id="password"
                type="password"
                placeholder="Enter your password"
                className="bg-neutral-800 border-neutral-700 text-white"
              />
            </div>
            <Button className="w-full bg-blue-600 hover:bg-blue-700">Sign in</Button>
          </form>
          <p className="mt-6 text-center text-sm text-neutral-400">
            Don’t have an account?{" "}
            <Link href="/auth/sign-up" className="hover:text-white">
              Sign up
            </Link>
          </p>
        </CardContent>
    </Card>
  );
}
