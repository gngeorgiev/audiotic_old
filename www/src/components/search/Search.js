import React, { Component } from 'react';
import './Search.css';

import { ServerUrl } from '../../constants';

class Search extends Component {
    state = {
        currentSuggestion: '',
        tracks: []
    }

    componentWillReceiveProps(nextProps) {
        if (nextProps.suggestion !== this.state.currentSuggestion) {
            this.currentSuggestion = nextProps.suggestion;
            this.search(nextProps.suggestion);
        }
    }

    search(suggestion) {
        suggestion = encodeURIComponent(suggestion)
        fetch(`${ServerUrl}/meta/search/${suggestion}`)
            .then(response => response.json())
            .then(tracks => this.setState({tracks}))
            .catch(err => console.error(err));
    }

    firePlayTrack(track) {
        this.props.playTrack(track);
    }

    render() {
        return (
            <div>
                {this.state.tracks.map((t, i) => (
                    <div key={i} className="search-item" onClick={() => this.firePlayTrack(t)}>
                        <img width="100" height="100" alt="thumbnail" src={t.thumbnail} />
                        <span>{t.title}</span>
                    </div>
                ))}
            </div>
        )
    }
}

export default Search;