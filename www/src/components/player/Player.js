import React, { Component } from 'react';
import './Player.css';

import { ServerUrl, DefaultThumbnail } from '../../constants';
import SockJS from 'sockjs-client'

import Button from 'react-md/lib/Buttons/Button';
import FontIcon from 'react-md/lib/FontIcons';

class Player extends Component {
    socket = null;

    state = {
        track: {},
        status: {
            length: 0,
            time: 0,
            name: '',
            state: 'stopped',
            thumbnail: '',
            isPlaying: false
        },
        volume: 100
    }

    componentDidMount() {
        this.connectToSocket();
    }

    componentWillUnmount() {
        this.onSocketDisconnected();
    }

    componentWillReceiveProps({ track }) {
        if (track.id) {
            this.setState({track});
            this.play(track);
        }
    }

    connectToSocket() {
        this.socket = new SockJS(`${ServerUrl}/player/updates`);
        this.socket.onmessage = ({ data }) => {
            const status = JSON.parse(data);
            this.updatePlayer(status);
        };
        this.socket.onclose = () => {
            this.onSocketDisconnected();
            setTimeout(() => this.connectToSocket(), 100);
        };
    }

    onSocketDisconnected() {
        this.socket.onmessage = null;
        this.socket = null;
    }

    renderPlayButton() {
        const isPlaying = this.state.status.isPlaying;
        let iconClassName = 'fa ';
        if (isPlaying) {
            iconClassName += 'fa-pause';
        } else {
            iconClassName += 'fa-play';
        }

        const playButtonClick = () => {
            if (isPlaying) {
                this.pause();
            } else {
                this.resume();
            }
        };

        return <Button 
            flat
            className="play-button"
            onClick={playButtonClick}
        >
            <FontIcon iconClassName={iconClassName} />
        </Button>
    }

    updatePlayer(status) {
        this.setState({
            status: {
                name: status.name,
                state: status.state,
                time: +status.time,
                length: +status.length,
                thumbnail: status.thumbnail,
                isPlaying: status.isPlaying
            }    
        });
    }

    play(track) {
        fetch(`${ServerUrl}/player/play/${track.provider}/${track.id}`)
            .catch(err => console.error(err));
    }

    pause() {
        fetch(`${ServerUrl}/player/pause`)
            .catch(err => console.error(err));
    }

    stop() {
        fetch(`${ServerUrl}/player/stop`)
            .catch(err => console.error(err));
    }

    resume() {
        fetch(`${ServerUrl}/player/resume`)
            .then(() => this.setState({playing: true}))
            .catch(err => console.error(err));
    }

    seek(time) {
        fetch(`${ServerUrl}/player/seek/${time}`)
            .catch(err => console.error(err));
    }

    render() {
        return (
            <footer>
                <div className="row">
                    <div className="col-xs-3 player-item">
                        <img className="thumbnail" alt="thumbnail" src={this.state.status.thumbnail || DefaultThumbnail} />
                    </div>
                    <div className="col-xs-7 player-item heading-container">
                        <h2 className="title">{this.state.status.name || 'Play something'}</h2>
                    </div>
                    <div className="col-xs-2 player-item">
                        {this.renderPlayButton()}
                    </div>
                </div>
            </footer>
        );
    }
}

export default Player;