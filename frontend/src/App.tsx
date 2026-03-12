import React, { useEffect, useState } from 'react';
import './App.css';

type Message = {
  ts: string;
  user: string;
  text: string;
};

type Thread = {
  channel_id: string;
  channel_name: string;
  thread_ts: string;
  messages: Message[];
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

function ThreadCard({ thread }: { thread: Thread }) {
  const [open, setOpen] = useState(false);
  const first = thread.messages[0];
  const replyCount = thread.messages.length - 1;

  return (
    <div className="thread-card">
      <div className="channel">#{thread.channel_name}</div>
      {first && (
        <div className="first-message">
          <span className="user">{first.user}</span>
          <div className="text"><SlackText text={first.text} /></div>
        </div>
      )}
      {replyCount > 0 && (
        <>
          <button className="toggle" onClick={() => setOpen(!open)}>
            {open ? '▼' : '▶'} {replyCount} {replyCount === 1 ? 'reply' : 'replies'}
          </button>
          {open && (
            <div className="replies">
              {thread.messages.slice(1).map((m) => (
                <div key={m.ts} className="reply">
                  <span className="user">{m.user}</span>
                  <div className="text"><SlackText text={m.text} /></div>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}

function App() {
  const [threads, setThreads] = useState<Thread[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch('http://localhost:8080/conversations')
      .then((res) => res.json())
      .then((data) => setThreads(data ?? []))
      .catch((err) => setError(err.message));
  }, []);

  if (error) {
    return <div className="App"><p>Error: {error}</p></div>;
  }

  return (
    <div className="App">
      <h1>Watcher</h1>
      <p>{threads.length} threads</p>
      <div className="threads">
        {threads.map((t) => (
          <ThreadCard key={`${t.channel_id}-${t.thread_ts}`} thread={t} />
        ))}
      </div>
    </div>
  );
}

export default App;
