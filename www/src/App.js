import React, { Component } from 'react';
import './App.css';

import Player from './components/player/Player';
import Search from './components/search/Search';

class App extends Component {
  render() {
    return (
      <div className="App">
        <div className="App-header">
          <Player />
        </div>
          <Search className="App-intro" />
      </div>
    );
  }
}

export default App;
