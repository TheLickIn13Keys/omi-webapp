"use client"
import React, { useState, useEffect, useRef } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Slider } from "@/components/ui/slider"
import { Switch } from "@/components/ui/switch"
import { Music, Play, Pause, Send } from 'lucide-react'
import { useToast } from "@/hooks/use-toast"


interface Conversation {
  id: string;
  name: string;
}

interface AudioFile {
  name: string;
  url: string;
}

interface TranscriptionWord {
  word: string;
  start: number;
  end: number;
  confidence: number;
}

interface TranscriptionSentence {
  sentence: string;
  start: number;
  end: number;
  words: TranscriptionWord[];
  confidence: number;
  speaker: string | null;
  channel: number;
}

export default function ConversationView({ conversation }: { conversation: Conversation }) {
  const [chatMessage, setChatMessage] = useState('')
  const [transcript, setTranscript] = useState<TranscriptionSentence[]>([])
  const [chatHistory, setChatHistory] = useState<string[]>([])
  const [audioFile, setAudioFile] = useState<AudioFile | null>(null)
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const [summary, setSummary] = useState<string>('')
  const [actionItems, setActionItems] = useState<string[]>([])
  const audioRef = useRef<HTMLAudioElement>(null)
  const { toast } = useToast()

  useEffect(() => {
    fetchConversationDetails()
    fetchAudioFile()
  }, [conversation.id])

  useEffect(() => {
    if (audioRef.current) {
      const audio = audioRef.current;
      audio.addEventListener('timeupdate', handleTimeUpdate);
      audio.addEventListener('loadedmetadata', handleLoadedMetadata);
      audio.addEventListener('ended', handleAudioEnded);
      audio.addEventListener('error', handleAudioError);

      return () => {
        audio.removeEventListener('timeupdate', handleTimeUpdate);
        audio.removeEventListener('loadedmetadata', handleLoadedMetadata);
        audio.removeEventListener('ended', handleAudioEnded);
        audio.removeEventListener('error', handleAudioError);
      };
    }
  }, [audioRef]);

  useEffect(() => {
    if (audioRef.current && audioFile) {
      audioRef.current.src = audioFile.url
      audioRef.current.load() 
      setError(null) 
      if (isPlaying) {
        audioRef.current.play().catch(handleAudioError)
      }
    }
  }, [audioFile, isPlaying])

  const fetchConversationDetails = async () => {
    try {
      const token = localStorage.getItem('token')
      const response = await fetch(process.env.BASE_URL + `/conversations/${conversation.id}`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (response.ok) {
        const data = await response.json()
        setTranscript(data.transcript || [])
        setChatHistory(data.chatHistory || [])
        setSummary(data.summary || '')
        setActionItems(data.action_items || [])
      } else {
        toast({
          title: "Error",
          description: "Failed to fetch conversation details",
          variant: "destructive",
        })
      }
    } catch (error) {
      console.error('Error fetching conversation details:', error)
      toast({
        title: "Error",
        description: "An error occurred while fetching conversation details",
        variant: "destructive",
      })
    }
  }

  const fetchAudioFile = async () => {
    try {
      const token = localStorage.getItem('token')
      const response = await fetch(process.env.BASE_URL + `/conversations/${conversation.id}/audio`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      const data = await response.json()
      console.log('Full API response:', data);
      console.log('Response status:', response.status);
      console.log('Response headers:', Object.fromEntries(response.headers.entries()));
      
      if (response.ok) {
        console.log('Received audio file:', data.audio_file);
        if (data.audio_file) {
          setAudioFile(data.audio_file)
        } else {
          console.warn('No audio file found for this conversation');
          toast({
            title: "Info",
            description: "No audio file found for this conversation",
            variant: "default",
          })
        }
      } else {
        throw new Error(data.error || 'Unknown error')
      }
    } catch (error) {
      console.error('Error fetching audio file:', error)
      toast({
        title: "Error",
        description: "An error occurred while fetching the audio file",
        variant: "destructive",
      })
    }
  }

  const handleSendMessage = async () => {
    if (!chatMessage.trim()) return

    try {
      const token = localStorage.getItem('token')
      const response = await fetch(process.env.BASE_URL + `/conversations/${conversation.id}/messages`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ message: chatMessage })
      })

      if (response.ok) {
        setChatHistory([...chatHistory, `You: ${chatMessage}`])
        setChatMessage('')
        
        
      } else {
        toast({
          title: "Error",
          description: "Failed to send message",
          variant: "destructive",
        })
      }
    } catch (error) {
      console.error('Error sending message:', error)
      toast({
        title: "Error",
        description: "An error occurred while sending the message",
        variant: "destructive",
      })
    }
  }

  const handlePlayPause = () => {
    if (audioRef.current) {
      if (isPlaying) {
        audioRef.current.pause()
      } else {
        audioRef.current.play().catch(handleAudioError)
      }
      setIsPlaying(!isPlaying)
    }
  }

  const handleTimeUpdate = () => {
    if (audioRef.current) {
      setCurrentTime(audioRef.current.currentTime)
    }
  }

  const handleLoadedMetadata = () => {
    if (audioRef.current) {
      setDuration(audioRef.current.duration)
    }
  }

  const handleAudioEnded = () => {
    setIsPlaying(false)
    setCurrentTime(0)
  }

  const handleAudioError = (e: Event) => {
    const target = e.target as HTMLAudioElement
    console.error('Audio error:', target.error)
    setError(`Error playing audio: ${target.error?.message || 'Unknown error'}`)
    setIsPlaying(false)
    toast({
      title: "Audio Error",
      description: `Failed to play audio: ${target.error?.message || 'Unknown error'}`,
      variant: "destructive",
    })
  }

  const handleSliderChange = (value: number[]) => {
    if (audioRef.current) {
      audioRef.current.currentTime = value[0]
      setCurrentTime(value[0])
    }
  }

  const formatTime = (time: number) => {
    const minutes = Math.floor(time / 60)
    const seconds = Math.floor(time % 60)
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }


  return (
    <>
      <h2 className="text-2xl font-bold mb-4">{conversation.name}</h2>
      <Card className="mb-4">
        <CardHeader>
          <CardTitle>Quick Info</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="summary">
            <TabsList>
              <TabsTrigger value="summary">Summary</TabsTrigger>
              <TabsTrigger value="action-items">Action Items</TabsTrigger>
            </TabsList>
            <TabsContent value="summary">
              <ScrollArea className="h-[200px]">
                <p>{summary || "No summary available"}</p>
              </ScrollArea>
            </TabsContent>
            <TabsContent value="action-items">
              <ScrollArea className="h-[200px]">
                {actionItems.length > 0 ? (
                  <ul className="list-disc pl-5">
                    {actionItems.map((item, index) => (
                      <li key={index}>{item}</li>
                    ))}
                  </ul>
                ) : (
                  <p>No action items available</p>
                )}
              </ScrollArea>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Audio Player */}
      <Card className="mb-4">
        <CardHeader>
          <CardTitle>Audio Player</CardTitle>
        </CardHeader>
        <CardContent>
          <audio
            ref={audioRef}
            onTimeUpdate={handleTimeUpdate}
            onLoadedMetadata={handleLoadedMetadata}
            onEnded={handleAudioEnded}
          />
          {error && (
            <div className="text-red-500 mb-2">{error}</div>
          )}
          {audioFile ? (
            <div className="space-y-4">
              <Slider
                value={[currentTime]}
                max={duration}
                step={0.1}
                onValueChange={handleSliderChange}
              />
              <div className="flex justify-between items-center">
                <span>{formatTime(currentTime)}</span>
                <Button size="icon" variant="outline" onClick={handlePlayPause}>
                  {isPlaying ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                </Button>
                <span>{formatTime(duration)}</span>
              </div>
              <div className="flex items-center space-x-2">
                <Music className="h-4 w-4" />
                <span>{audioFile.name}</span>
              </div>
            </div>
          ) : (
            <div className="text-center text-gray-500">No audio file available for this conversation</div>
          )}
        </CardContent>
      </Card>

      {/* Transcript and Chat */}
      <Tabs defaultValue="transcript" className="space-y-4">
        <TabsList>
          <TabsTrigger value="transcript">Transcript</TabsTrigger>
          <TabsTrigger value="chat">Chat</TabsTrigger>
        </TabsList>
        <TabsContent value="transcript" className="space-y-4">
          <Card>
            <CardContent className="p-4">
              <ScrollArea className="h-[300px]">
                <div className="space-y-4">
                  {transcript.map((sentence, index) => (
                    <div key={index} className="mb-2">
                      <p className="text-sm text-gray-500">
                        {formatTime(sentence.start)} - {formatTime(sentence.end)}
                        {sentence.speaker && ` | Speaker: ${sentence.speaker}`}
                      </p>
                      <p>{sentence.sentence}</p>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="chat" className="space-y-4">
          <Card>
            <CardContent className="p-4">
              <ScrollArea className="h-[300px]">
                <div className="space-y-4">
                  <p><strong>You:</strong> What was the main topic of the conversation?</p>
                  <p><strong>AI:</strong> The main topic of the conversation was a friendly greeting and exchange of pleasantries.</p>
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
          <div className="flex space-x-2">
            <Input
              type="text"
              placeholder="Ask about your recordings..."
              value={chatMessage}
              onChange={(e) => setChatMessage(e.target.value)}
            />
            <Button size="icon">
              <Send className="h-4 w-4" />
            </Button>
          </div>
        </TabsContent>
      </Tabs>

      {/* Plugin Outputs */}
      <Card className="mt-4">
        <CardHeader>
          <CardTitle>Plugin Outputs</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">Sentiment Analysis</CardTitle>
              </CardHeader>
              <CardContent>
                <p>Positive: 80%</p>
                <p>Neutral: 15%</p>
                <p>Negative: 5%</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">Bias Detection</CardTitle>
              </CardHeader>
              <CardContent>
                <p>No significant bias detected</p>
              </CardContent>
            </Card>
          </div>
        </CardContent>
      </Card>

      {/* Music Analyzer */}
      <Card className="mt-4">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Music Analyzer</CardTitle>
          <Switch />
        </CardHeader>
        <CardContent>
          <div className="flex items-center space-x-4">
            <Music className="h-4 w-4" />
            <div>
              <p className="text-sm font-medium">Detected Song</p>
              <p className="text-xs text-muted-foreground">Artist - Song Name</p>
            </div>
            <span className="text-xs text-muted-foreground">00:15 - 00:45</span>
          </div>
        </CardContent>
      </Card>
    </>
  )
}