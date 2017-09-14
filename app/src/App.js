import React, { Component } from 'react';

import Message from './components/Message';
import './App.css';

const socket = new WebSocket("ws://127.0.0.1:8081/ws");
socket.onopen = (event) => {
  console.log('Opened socket');
};

socket.onerror = (event) => {
  console.log(event);
}

class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      init: true,
      participants: [],
      messages: [],
    };

    this.participantSubmit = this.participantSubmit.bind(this);
    this.textSubmit = this.textSubmit.bind(this);
  }

  participantSubmit(e) {
    e.preventDefault();
    const input = document.getElementById("participant-form").value;
    console.log(input);
    if (input === "") {
      return;
    }

    this.setState({
      init: false,
      participants: [
        ...this.state.participants,
        input,
      ]
    });
  }

  textSubmit(e) {
    e.preventDefault();
    const input = document.getElementById("text-entry-input").value;
    if (input === "") {
      return;
    }

    socket.send(JSON.stringify({
      text: input,
      client: "testClient",
      timestamp: new Date().toLocaleTimeString(),
    }));

    document.getElementById("text-entry-input").value = "";
  }

  componentDidUpdate() {
    // keep message box scrolled appropriately with new messages
    if (!this.state.init) {
      document.getElementById("conversation-main").scrollTop = document.getElementById("conversation-main").scrollHeight;
    }
  }

  componentDidMount() {
    socket.onmessage = (event) => {
      const { text, client, timestamp } = JSON.parse(event.data);

      this.setState({
        messages: [
          ...this.state.messages,
          {
            text,
            client,
            timestamp,
          }
        ],
      });
    }
  }

  render() {
    console.log(this.state);
    return (
      <div id="box-main">
        {this.state.init &&
          <div id="init">
            <form onSubmit={this.participantSubmit}>
              <input type="text"
                     placeholder="Enter name here..."
                     id="participant-form" />
            </form>
          </div>
        }
        {!this.state.init &&
          <div style={{width: "100%", height: "100%"}}>
            <div id="top">
              <div id="participants-main">
                <b>{`Participants (${this.state.participants.length}):`}</b>
                {this.state.participants.map(participant =>
                  <div>{participant}</div>
                )}
              </div>
              <div id="conversation-main">
                <a href="https://github.com/stefaluc/go-react-chat"
                  id="source"
                  target="_blank"
                  rel="noopener noreferrer">View source code</a>
                {this.state.messages.map(message =>
                  <Message message={message} />
                )}
              </div>
            </div>
            <div id="bottom">
              <div id="text-entry-main">
                <form onSubmit={this.textSubmit} id="text-entry-form">
                  <input type="text" id="text-entry-input" />
                </form>
              </div>
            </div>
          </div>
        }
      </div>
    );
  }
}

export default App;
