import { useState } from 'react'
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { ChevronLeft, Search } from 'lucide-react'

interface PluginsMarketplaceProps {
  onClose: () => void;
}

export default function PluginsMarketplace({ onClose }: PluginsMarketplaceProps) {
  const [searchQuery, setSearchQuery] = useState('')

  const plugins = [
    { id: 1, name: 'Sentiment Analyzer', description: 'Analyze the sentiment of your conversations', installed: true },
    { id: 2, name: 'Bias Detector', description: 'Detect potential biases in speech', installed: true },
    { id: 3, name: 'Topic Extractor', description: 'Extract main topics from your conversations', installed: false },
    { id: 4, name: 'Language Translator', description: 'Translate conversations to multiple languages', installed: false },
    { id: 5, name: 'Emotion Detector', description: 'Detect emotions in speech', installed: false },
    { id: 6, name: 'Keyword Highlighter', description: 'Highlight important keywords in transcripts', installed: false },
  ]

  const filteredPlugins = plugins.filter(plugin => 
    plugin.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    plugin.description.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Plugins Marketplace</h1>
        <Button variant="outline" onClick={onClose}>
          <ChevronLeft className="mr-2 h-4 w-4" /> Back to Dashboard
        </Button>
      </div>

      <div className="flex space-x-2">
        <Input
          type="text"
          placeholder="Search plugins..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="flex-grow"
        />
        <Button size="icon">
          <Search className="h-4 w-4" />
        </Button>
      </div>

      <ScrollArea className="h-[calc(100vh-200px)]">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filteredPlugins.map(plugin => (
            <Card key={plugin.id}>
              <CardHeader>
                <CardTitle>{plugin.name}</CardTitle>
                <CardDescription>{plugin.description}</CardDescription>
              </CardHeader>
              <CardFooter>
                <Button variant={plugin.installed ? "secondary" : "default"}>
                  {plugin.installed ? "Installed" : "Install"}
                </Button>
              </CardFooter>
            </Card>
          ))}
        </div>
      </ScrollArea>
    </div>
  )
}