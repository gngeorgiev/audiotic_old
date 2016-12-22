import React, { Component, PropTypes } from 'react';
import './Header.css';

import { ServerUrl } from '../../constants';
import Toolbar from 'react-md/lib/Toolbars';
import Autocomplete from 'react-md/lib/Autocompletes';
import Button from 'react-md/lib/Buttons/Button';
import { throttle } from 'lodash';

class Header extends Component {
    static propTypes = {
        suggestionSelected: PropTypes.func
    }

    state = {
        suggestions: [],
        value: '',
        searching: false,
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

    back() {
        this.setState({searching: false});
        this.searchSuggestion('');
    }

    render() {
        const { suggestions, value, searching } = this.state;

        let nav;
        let title;
        let actions;
        let children;
        if (searching) {
            nav = <Button icon onClick={() => this.back()}>arrow_back</Button>
            actions = <Button icon onClick={this.reset.bind(this)}>close</Button>
            children = (
                <Autocomplete
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
                    className="md-title--toolbar md-cell autocomplete"
                    inputClassName="md-text-field--toolbar"
                />
            );
        } else {
            nav = <Button icon>history</Button>;;
            title = 'History';
            actions = <Button icon onClick={() => this.setState({searching: true})}>search</Button>
        }

        return (
            <Toolbar
                colored
                inset
                fixed
                className="md-paper md-paper--1 toolbar"
                actions={actions}
                title={title}
                nav={nav}
            >
                {children}
            </Toolbar>
        );
    }
}

export default Header;