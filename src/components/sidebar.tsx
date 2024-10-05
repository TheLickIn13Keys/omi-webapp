import React, { useState, useRef } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { ChevronRight, LogOut, Send, Upload } from 'lucide-react'
import { useAuth } from '@/contexts/AuthContext'
import { useRouter } from 'next/navigation'
import { useToast } from "@/hooks/use-toast"

interface TranscriptionSentence {
  sentence: string;
  start: number;
  end: number;
  words: Array<{
    word: string;
    start: number;
    end: number;
    confidence: number;
  }>;
  confidence: number;
  speaker: string | null;
  channel: number;
}

interface Conversation {
  id: string;
  name: string;
  transcript: TranscriptionSentence[];
  summary: string;
  actionItems: string[];
}

interface SidebarProps {
  conversations: Conversation[] | null;
  onConversationSelect: (conversation: Conversation) => void;
  onPluginsClick: () => void;
}

export default function Sidebar({ conversations, onConversationSelect, onPluginsClick }: SidebarProps) {
  const [chatMessage, setChatMessage] = useState('')
  const [isUploading, setIsUploading] = useState(false)
  const { logout } = useAuth()
  const router = useRouter()
  const { toast } = useToast()
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleSendMessage = async () => {
    // Implement sending message to backend
    console.log('Sending message:', chatMessage)
    setChatMessage('')
  }

  const handleLogout = async () => {
    try {
      await logout()
      router.push('/login') // Redirect to login page after logout
    } catch (error) {
      console.error('Logout failed:', error)
      // Handle logout error (show a message to the user, etc.)
    }
  }

  const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    setIsUploading(true)
    const formData = new FormData()
    formData.append('file', file)

    try {
      const token = localStorage.getItem('token')
      const response = await fetch("https://aggieworks-backend.server.bardia.app" + '/upload-audio', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData
      })

      if (response.ok) {
        const data = await response.json()
        toast({
          title: "Success",
          description: "File uploaded successfully",
        })
        // You might want to refresh the conversation list here
      } else {
        throw new Error('File upload failed')
      }
    } catch (error) {
      console.error('Error uploading file:', error)
      toast({
        title: "Error",
        description: "Failed to upload file",
        variant: "destructive",
      })
    } finally {
      setIsUploading(false)
    }
  }

  return (
    <div className="w-64 bg-white p-4 flex flex-col">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Omi Friend</h1>
        <Button variant="ghost" size="icon" onClick={handleLogout}>
          <LogOut className="h-5 w-5" />
        </Button>
      </div>
      <div className="flex space-x-2 mb-4">
        <Input
          type="text"
          placeholder="Ask me anything..."
          value={chatMessage}
          onChange={(e) => setChatMessage(e.target.value)}
          className="flex-grow"
        />
        <Button size="icon" onClick={handleSendMessage}>
          <Send className="h-4 w-4" />
        </Button>
      </div>
      <ScrollArea className="flex-grow">
        <div className="space-y-2">
          {conversations && conversations.length > 0 ? (
            conversations.map((conversation) => (
              <Button
                key={conversation.id}
                variant="ghost"
                className="w-full justify-start"
                onClick={() => onConversationSelect(conversation)}
              >
                <ChevronRight className="mr-2 h-4 w-4" /> {conversation.name}
              </Button>
            ))
          ) : (
            <p className="text-gray-500 text-center">No conversations yet</p>
          )}
        </div>
      </ScrollArea>
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFileUpload}
        style={{ display: 'none' }}
        accept="audio/*"
      />
      <Button 
        variant="outline" 
        className="w-full mt-4" 
        onClick={() => fileInputRef.current?.click()}
        disabled={isUploading}
      >
        {isUploading ? 'Uploading...' : 'Upload Audio File'}
        <Upload className="ml-2 h-4 w-4" />
      </Button>
      <Button variant="outline" className="w-full mt-4" onClick={onPluginsClick}>
        Plugins Marketplace
      </Button>
    </div>
  )
}