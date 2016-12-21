import React, { Component } from 'react';
import './App.css';

import WebFontLoader from 'webfontloader';

import Player from './components/player/Player';
import Suggest from './components/suggest/Suggest';
import Search from './components/search/Search';

import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';

WebFontLoader.load({
  google: {
    families: ['Roboto:300,400,500,700', 'Material Icons'],
  },
});

import injectTapEventPlugin from 'react-tap-event-plugin';

// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin();

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
      <MuiThemeProvider>
        <div>
          <Suggest className="App-intro" suggestionSelected={(suggestion) => this.setState({suggestion})} />
          <Search suggestion={suggestion} playTrack={this.playTrack.bind(this)}/>
          <Player track={track} />
        </div>
      </MuiThemeProvider>
    );
  }
}

export default App;
