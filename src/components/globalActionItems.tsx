"use client"


import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ScrollArea } from "@/components/ui/scroll-area"

export default function ActionItemsSummary() {
  const actionItems = [
    { id: 1, date: '2023-06-01', text: 'Follow up with client about project timeline' },
    { id: 2, date: '2023-06-01', text: 'Prepare slides for team meeting' },
    { id: 3, date: '2023-05-31', text: 'Review and approve budget proposal' },
    { id: 4, date: '2023-05-31', text: 'Schedule interview with potential hire' },
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle>Quick Info</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="action-items">
          <TabsList>
            <TabsTrigger value="action-items">Action Items</TabsTrigger>
            <TabsTrigger value="summary">Summary</TabsTrigger>
          </TabsList>
          <TabsContent value="action-items">
            <ScrollArea className="h-[200px]">
              {actionItems.map((item) => (
                <div key={item.id} className="py-2">
                  <p className="text-sm text-gray-500">{item.date}</p>
                  <p>{item.text}</p>
                </div>
              ))}
            </ScrollArea>
          </TabsContent>
          <TabsContent value="summary">
            <p>Today's summary: 2 meetings attended, 3 action items created, 1 project milestone reached.</p>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}