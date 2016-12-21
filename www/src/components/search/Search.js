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
        tracks: []
    }

    componentWillReceiveProps(nextProps) {
        if (nextProps.suggestion !== this.state.currentSuggestion) {
            this.setState({
                currentSuggestion: nextProps.suggestion
            });
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
            <div className="search">
                <List>
                    <ListItem primaryText=""/>
                    {this.state.currentSuggestion ? <Subheader primaryText={`Results for "${this.state.currentSuggestion}"`} /> : null}
                    {this.state.tracks.map((t, i) => (
                        <ListItem
                            key={i}
                            onClick={() => this.firePlayTrack(t)}
                            leftAvatar={<Avatar src={t.thumbnail} alt="thumbnail" />}
                            primaryText={t.title}
                        />
                    ))}
                </List>
            </div>
        )
    }
}

export default Search;