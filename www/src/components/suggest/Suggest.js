import React, { Component, PropTypes } from 'react';
import './Suggest.css';

import { ServerUrl } from '../../constants';
import Toolbar from 'react-md/lib/Toolbars';
import Autocomplete from 'react-md/lib/Autocompletes';
import Button from 'react-md/lib/Buttons/Button';
import { throttle } from 'lodash';

class Suggest extends Component {
    static propTypes = {
        suggestionSelected: PropTypes.func
    }

    state = {
        suggestions: [],
        value: ''
    }

    constructor() {
        super();

        this.throttleSuggest = throttle((value) => {
            fetch(`${ServerUrl}/meta/autocomplete/${value}`)
                .then(response => response.json())
                .then(suggestions => this.setState({suggestions}))
                .catch(err => console.error(err));
        }, 400);
    }

    reset() {
        this.setState({value: ''});
    }

    suggest(value) {
        this.setState({value})
        this.throttleSuggest(value);
    }

    searchSuggestion(suggestion) {
        this.props.suggestionSelected(suggestion);
    }

    onKeyDown(ev) {
        if (ev.key === 'Enter') {
            this.searchSuggestion(this.state.value);
            this.refs.autocomplete._close();
        }
    }

    render() {
        const { suggestions, value } = this.state;

        return (
            <Toolbar
                inset={true}
                fixed
                themed
                className="md-paper md-paper--1 toolbar"
                actions={<Button icon onClick={this.reset.bind(this)}>close</Button>}
                nav={<Button icon>search</Button>}
            >
                <Autocomplete
                    id="spotify-search"
                    type="search"
                    ref="autocomplete"
                    placeholder="Search tracks"
                    data={suggestions}
                    value={value}
                    filter={null}
                    onChange={this.suggest.bind(this)}
                    onAutocomplete={this.searchSuggestion.bind(this)}
                    onKeyDown={this.onKeyDown.bind(this)}
                    block
                    className="md-title--toolbar md-cell"
                    inputClassName="md-text-field--toolbar"
                />
            </Toolbar>
        );
    }
}

export default Suggest;