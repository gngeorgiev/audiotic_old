import React, { Component } from 'react';
import './App.css';

import WebFontLoader from 'webfontloader';

import Player from './components/player/Player';
import Suggest from './components/suggest/Suggest';
import Search from './components/search/Search';

WebFontLoader.load({
  google: {
    families: ['Roboto:300,400,500,700', 'Material Icons'],
  },
});

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
      <div>
        <Suggest className="App-intro" suggestionSelected={(suggestion) => this.setState({suggestion})} />
        <Search suggestion={suggestion} playTrack={this.playTrack.bind(this)}/>
        <Player track={track} />
      </div>
    );
  }
}

export default App;
