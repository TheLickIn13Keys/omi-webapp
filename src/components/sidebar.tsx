import React, { useState } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { ChevronRight, LogOut, Send } from 'lucide-react'
import { useAuth } from '@/contexts/AuthContext'
import { useRouter } from 'next/navigation'

interface Conversation {
  id: string;
  name: string;
}

interface SidebarProps {
  conversations: Conversation[] | null;
  onConversationSelect: (conversation: Conversation) => void;
  onPluginsClick: () => void;
}

export default function Sidebar({ conversations, onConversationSelect, onPluginsClick }: SidebarProps) {
  const [chatMessage, setChatMessage] = useState('')
  const { logout } = useAuth()
  const router = useRouter()

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
      <Button variant="outline" className="w-full mt-4" onClick={onPluginsClick}>
        Plugins Marketplace
      </Button>
    </div>
  )
}