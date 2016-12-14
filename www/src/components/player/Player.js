import React, { Component } from 'react';
import './Player.css';

import { Button } from 'react-bootstrap';

class Player extends Component {
    state = {
        playing: false,
        title: 'test title'
    }

    renderPlayButton() {
        let className = 'col-md-1 fa ';
        if (this.state.playing) {
            className += 'fa-pause-circle';
        } else {
            className += 'fa-play-circle';
        }

        return <Button className={className} />
    }

    render() {
        return (
            <div className="player">
                <h4>{this.state.title}</h4>
                <div className="row">
                    {this.renderPlayButton()}
                </div>
            </div>
        );
    }
}

export default Player;