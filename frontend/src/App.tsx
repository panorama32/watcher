import React, { useEffect, useState } from 'react';
import './App.css';

type Message = {
  channel_id: string;
  channel_name: string;
  ts: string;
  user: string;
  text: string;
};

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
            <span className="channel">#{m.channel_name}</span>
            <span className="user">{m.user}</span>
            <span className="text">{m.text}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
