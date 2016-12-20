import React, { Component } from 'react';
import './Player.css';

import { Button } from 'react-bootstrap';
import { ServerUrl, DefaultThumbnail } from '../../constants';
import SockJS from 'sockjs-client'
import 'rc-slider/assets/index.css';
import Slider from 'rc-slider';

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
        let className = 'col-md-1 fa ';
        const isPlaying = this.state.status.isPlaying;
        if (isPlaying) {
            className += 'fa-pause';
        } else {
            className += 'fa-play';
        }

        return <Button className={className} onClick={() => {
            if (isPlaying) {
                this.pause();
            } else {
                this.resume();
            }
        }} />
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
            <div className="player" style={{
                backgroundImage: this.state.status.thumbnail   
            }}>
                <img className="thumbnail" alt="thumbnail" src={this.state.status.thumbnail || DefaultThumbnail} />
                <h4>{this.state.status.name}</h4>
                <div className="row buttons-row">
                    <div className="col-md-4 col-md-offset-4">
                        {this.renderPlayButton()}
                        <Button className="col-md-1 fa fa-stop" onClick={this.stop.bind(this)} />
                    </div>
                </div>
                <div className="row">
                    <div className="col-md-4 col-md-offset-4">
                        <Slider
                            className="seek-slider"
                            value={this.state.status.time} 
                            defaultValue={0} 
                            min={0} 
                            max={100} 
                            onAfterChange={this.seek.bind(this)} 
                        />
                    </div>
                </div>
            </div>
        );
    }
}

export default Player;