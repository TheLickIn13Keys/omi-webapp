"use client";
import React, { useState, useEffect } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import { useRouter } from 'next/navigation'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card } from "@/components/ui/card"
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog"
import Sidebar from './sidebar'
import ActionItemsSummary from './globalActionItems'
import ConversationView from './conversationView'
import SettingsModal from './settingsModal'
import PluginsMarketplace from './plugins'
import { Cog, Search, RefreshCw } from 'lucide-react'
import { useToast } from "@/hooks/use-toast"

type Conversation = {
  id: string;
  name: string;
};

export default function Dashboard() {
  const [globalSearch, setGlobalSearch] = useState('')
  const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null)
  const [conversations, setConversations] = useState<Conversation[] | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [showPluginsMarketplace, setShowPluginsMarketplace] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const { isAuthenticated } = useAuth()
  const router = useRouter()
  const { toast } = useToast()

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login')
    } else {
      fetchConversations()
    }
  }, [isAuthenticated, router])

  const fetchConversations = async () => {
    setIsLoading(true)
    try {
      const token = localStorage.getItem('token')
      const response = await fetch('http://localhost:8080/conversations', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (response.ok) {
        const data = await response.json()
        setConversations(data)
      } else {
        console.error('Failed to fetch conversations')
        toast({
          title: "Error",
          description: "Failed to fetch conversations",
          variant: "destructive",
        })
      }
    } catch (error) {
      console.error('Error fetching conversations:', error)
      toast({
        title: "Error",
        description: "An error occurred while fetching conversations",
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  const refreshConversations = async () => {
    setIsRefreshing(true)
    try {
      const token = localStorage.getItem('token')
      const response = await fetch('http://localhost:8080/query-bucket', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (response.ok) {
        const data = await response.json()
        if (data.new_conversations && data.new_conversations.length > 0) {
          toast({
            title: "Success",
            description: `Found ${data.new_conversations.length} new conversation(s)`,
          })
          await fetchConversations() // Refresh the conversation list
        } else {
          toast({
            title: "Info",
            description: "No new conversations found",
          })
        }
      } else {
        console.error('Failed to query bucket')
        toast({
          title: "Error",
          description: "Failed to query bucket",
          variant: "destructive",
        })
      }
    } catch (error) {
      console.error('Error querying bucket:', error)
      toast({
        title: "Error",
        description: "An error occurred while querying the bucket",
        variant: "destructive",
      })
    } finally {
      setIsRefreshing(false)
    }
  }

  const togglePluginsMarketplace = () => {
    setShowPluginsMarketplace(!showPluginsMarketplace)
  }

  if (!isAuthenticated) {
    return null // or a loading spinner
  }

  return (
    <div className="flex h-screen bg-gray-100">
      <Sidebar
        conversations={conversations}
        onConversationSelect={setSelectedConversation}
        onPluginsClick={togglePluginsMarketplace}
      />
      <div className="flex-1 p-6 space-y-6 overflow-auto">
        {showPluginsMarketplace ? (
          <PluginsMarketplace onClose={togglePluginsMarketplace} />
        ) : (
          <>
            <div className="flex space-x-2">
              <Input
                type="text"
                placeholder="Search all transcripts..."
                value={globalSearch}
                onChange={(e) => setGlobalSearch(e.target.value)}
                className="flex-grow"
              />
              <Button size="icon">
                <Search className="h-4 w-4" />
              </Button>
              <Button size="icon" onClick={refreshConversations} disabled={isRefreshing}>
                <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
              </Button>
            </div>

            <ActionItemsSummary />

            {isLoading ? (
              <div className="flex items-center justify-center h-[calc(100vh-200px)]">
                <p className="text-xl text-gray-500">Loading conversations...</p>
              </div>
            ) : selectedConversation ? (
              <ConversationView conversation={selectedConversation} />
            ) : (
              <div className="flex items-center justify-center h-[calc(100vh-200px)]">
                <p className="text-xl text-gray-500">Select a conversation or start a new one</p>
              </div>
            )}
          </>
        )}
      </div>
      <Dialog>
        <DialogTrigger asChild>
          <Button variant="ghost" size="icon" className="absolute bottom-4 right-4">
            <Cog className="h-5 w-5" />
          </Button>
        </DialogTrigger>
        <DialogContent className="sm:max-w-[425px]">
          <SettingsModal />
        </DialogContent>
      </Dialog>
    </div>
  )
}