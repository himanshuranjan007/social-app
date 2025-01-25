"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"


// Mock function to get the logged-in user's email from a JWT token
const getUserEmail = () => {
  // const token = localStorage.getItem("authToken")
  // if (!token) {
  //   alert("No auth token found. Please log in.")
  //   window.location.href = "/login"
  // }
  // const decodedToken: { email: string } = jwtDecode(token as string)
  return "user@example"
}

interface Post {
  id: number
  content: string
  author: string
}

// Fetch posts from the backend
const fetchPosts = async (): Promise<Post[]> => {
  const response = await fetch("http://localhost:8080/getposts")
  if (!response.ok) {
    throw new Error("Failed to fetch posts")
  }
  const text = await response.text()
  return text ? JSON.parse(text) : {}
}

// Create a new post in the backend
const createPost = async (content: string, author: string): Promise<Post> => {
  const response = await fetch("http://localhost:8080/createpost", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ Content: content, Author: author }),
  })
  if (!response.ok) {
    throw new Error("Failed to create post")
  }
  return response.json()
}

export default function Home() {
  const [posts, setPosts] = useState<Post[]>([])
  const [newPost, setNewPost] = useState("")
  const userEmail = getUserEmail()

  useEffect(() => {
    fetchPosts().then(setPosts).catch(console.error)
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const post = await createPost(newPost, userEmail)
      setPosts([...posts, post])
      setNewPost("")
    } catch (error) {
      console.error(error)
    }
  }

  return (
    <div className="min-h-screen bg-gray-100 p-4">
      <div className="max-w-2xl mx-auto space-y-8">
        <h1 className="text-3xl font-bold">Home</h1>

        <form onSubmit={handleSubmit} className="space-y-4">
          <Textarea
            value={newPost}
            onChange={(e) => setNewPost(e.target.value)}
            placeholder="What's happening?"
            className="w-full"
          />
          <Button type="submit">Post</Button>
        </form>

        <div className="space-y-4">
          {posts && posts.map((post) => (
            <div key={post.id} className="bg-white p-4 rounded-lg shadow">
              <p>{post.content}</p>
              <p className="text-sm text-gray-500 mt-2">Posted by {post.author}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
