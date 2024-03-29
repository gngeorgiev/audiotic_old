import React, { Component } from 'react';
import './Search.css';

import { ServerUrl } from '../../constants';

import List from 'react-md/lib/Lists/List';
import ListItem from 'react-md/lib/Lists/ListItem';
import Avatar from 'react-md/lib/Avatars';
import Subheader from 'react-md/lib/Subheaders';

class Search extends Component {
    state = {
        currentSuggestion: '',
        tracks: [],
        history: []
    }

    componentWillReceiveProps(nextProps) {
        if (nextProps.suggestion !== this.state.currentSuggestion) {
            this.setState({
                currentSuggestion: nextProps.suggestion
            });
            this.search(nextProps.suggestion);
        }
    }

    componentDidMount() {
        this.fetchHistory();
    }

    search(suggestion) {
        if (!suggestion) {
            this.fetchHistory(); //we are coming back from search
            return;
        }

        suggestion = encodeURIComponent(suggestion)
        fetch(`${ServerUrl}/meta/search/${suggestion}`)
            .then(response => response.json())
            .then(tracks => this.setState({tracks}))
            .catch(err => console.error(err));
    }

    fetchHistory() {
        fetch(`${ServerUrl}/history/get`)
            .then(response => response.json())
            .then(history => this.setState({history}))
            .catch(err => console.error(err));
    }

    firePlayTrack(track) {
        this.props.playTrack(track);
        this.fetchHistory();
    }

    renderTracks(tracks) {
        return tracks.map((t, i) => (
            <ListItem
                key={i}
                onClick={() => this.firePlayTrack(t)}
                leftAvatar={<Avatar src={t.thumbnail} alt="thumbnail" />}
                primaryText={t.title}
            />
        ));
    }

    renderSearch() {
        return (
            <List>
                {this.state.currentSuggestion ? <Subheader primaryText={`Results for "${this.state.currentSuggestion}"`} /> : null}
                {this.renderTracks(this.state.tracks)}
            </List>
        )
    }

    renderHistory() {
        return (
            <List>
                {this.renderTracks(this.state.history)}
            </List>
        )
    }

    render() {
        return (
            <div className="search">
                {this.state.currentSuggestion ? 
                    this.renderSearch() :
                    this.renderHistory()}
            </div>
        )
    }
}

export default Search;