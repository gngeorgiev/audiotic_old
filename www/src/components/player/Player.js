import React, { Component } from 'react';
import './Player.css';

import { Button } from 'react-bootstrap';
import { ServerUrl } from '../../constants';

class Player extends Component {
    state = {
        track: {},
        status: {
            length: 0,
            time: 0,
            name: '',
            state: 'stopped',
            thumbnail: '',
        }
    }

    componentDidMount() {
        setInterval(() => {
            this.updatePlayer();
        }, 1000);
    }

    componentWillReceiveProps({ track }) {
        if (track.id) {
            this.setState({track});
            this.play(track);
        }
    }

    renderPlayButton() {
        let className = 'col-md-1 fa ';
        const isPlaying = this.state.status.state === 'playing' 
        if (isPlaying) {
            className += 'fa-pause-circle';
        } else {
            className += 'fa-play-circle';
        }

        return <Button className={className} onClick={() => {
            if (isPlaying) {
                this.pause();
            } else {
                this.resume();
            }
        }} />
    }

    updatePlayer() {
        fetch(`${ServerUrl}/player/status`)
            .then(response => response.json())
            .then(status => this.setState({
                status: {
                    name: status.name,
                    state: status.state,
                    time: +status.time,
                    length: +status.length
                }    
            }))
            .catch(err => console.error(err))
    }

    play(track) {
        fetch(`${ServerUrl}/player/play/${track.provider}/${track.id}`)
            .then(() => this.setState({playing: true}))
            .catch(err => console.error(err));
    }

    pause() {
        fetch(`${ServerUrl}/player/pause`)
            .then(() => this.setState({playing: false}))
            .catch(err => console.error(err));
    }

    resume() {
        fetch(`${ServerUrl}/player/resume`)
            .then(() => this.setState({playing: true}))
            .catch(err => console.error(err));
    }

    render() {
        return (
            <div className="player">
                {
                    this.state.status.thumbnail !== '' ? 
                        <div></div> :
                        <img className="thumbnail" alt="thumbnail" src={this.state.status.thumbnail} /> 
                }
                <h4>{this.state.status.name}</h4>
                <div className="row">
                    {this.renderPlayButton()}
                </div>
            </div>
        );
    }
}

export default Player;