import Link from "next/link"
import { Button } from "@/components/ui/button"

export default function Home() {
  return (
    <div className="min-h-screen bg-gray-100 flex flex-col justify-center items-center p-4">
      <h1 className="text-4xl font-bold mb-8">Welcome to Our Social Media App</h1>
      <div className="space-y-4">
        <Button asChild className="w-full">
          <Link href="/signin">Sign In</Link>
        </Button>
        <Button asChild variant="outline" className="w-full">
          <Link href="/signup">Sign Up</Link>
        </Button>
      </div>
    </div>
  )
}

