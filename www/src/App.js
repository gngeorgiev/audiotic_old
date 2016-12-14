import React, { Component } from 'react';
import './App.css';

import Player from './components/player/Player';
import Suggest from './components/suggest/Suggest';
import Search from './components/search/Search';

class App extends Component {
  state = {
    suggestion: '',
    track: {}
  }

  playTrack(track) {
    this.setState({track});
  }

  render() {
    const { suggestion, track } = this.state;

    return (
      <div className="App">
        <div className="App-header">
          <Player track={track} />
        </div>
        <Suggest className="App-intro" suggestionSelected={(suggestion) => this.setState({suggestion})} />
        <Search suggestion={suggestion} playTrack={this.playTrack.bind(this)}/>
      </div>
    );
  }
}

export default App;
