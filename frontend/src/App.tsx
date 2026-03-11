import React, { useEffect, useState } from 'react';
import './App.css';

type Message = {
  channel_id: string;
  channel_name: string;
  ts: string;
  user: string;
  text: string;
};

function SlackText({ text }: { text: string }) {
  const parts = text.split(/(<[^>]+>)/g);
  return (
    <>
      {parts.map((part, i) => {
        const linkMatch = part.match(/^<(https?:\/\/[^|>]+)\|([^>]+)>$/);
        if (linkMatch) {
          return <a key={i} href={linkMatch[1]} target="_blank" rel="noopener noreferrer">{linkMatch[2]}</a>;
        }
        const urlMatch = part.match(/^<(https?:\/\/[^>]+)>$/);
        if (urlMatch) {
          return <a key={i} href={urlMatch[1]} target="_blank" rel="noopener noreferrer">{urlMatch[1]}</a>;
        }
        return <span key={i}>{part}</span>;
      })}
    </>
  );
}

function App() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch('http://localhost:8080/conversations')
      .then((res) => res.json())
      .then((data) => setMessages(data ?? []))
      .catch((err) => setError(err.message));
  }, []);

  if (error) {
    return <div className="App"><p>Error: {error}</p></div>;
  }

  return (
    <div className="App">
      <h1>Watcher</h1>
      <p>{messages.length} messages</p>
      <div className="conversations">
        {messages.map((m) => (
          <div key={`${m.channel_id}-${m.ts}`} className="message">
            <div className="channel">#{m.channel_name}</div>
            <div className="user">{m.user}</div>
            <div className="text"><SlackText text={m.text} /></div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
