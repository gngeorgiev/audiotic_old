import React, { Component } from 'react';
import './Search.css';

import Autosuggest from 'react-autosuggest';
import { ServerUrl } from '../../constants';

class Search extends Component {
    state = {
        suggestions: [],
        value: ''
    }

    onChange(event, { newValue }) {
        this.setState({value: newValue});
    }
    
    renderSuggestion(suggestion) {
        return <span>{suggestion}</span>
    }

    getSuggestionValue(suggestion) {
        return suggestion;
    }

    onSuggestionsFetchRequested(text) {
        fetch(`${ServerUrl}/meta/autocomplete/${text}`)
            .then(response => response.json())
            .then(suggestions => this.renderState({suggestions}))
            .catch(err => console.error(err));
    }

    onSuggestionsClearRequested() {
        this.setState({suggestions: []});
    }

    render() {
        const { suggestions, value } = this.state;

        return (
            <Autosuggest
                inputProps={{
                    placeholder: 'Search for tracks',
                    onChange: this.onChange.bind(this),
                    value
                }}

                suggestions={suggestions}
                onSuggestionsFetchRequested={this.onSuggestionsFetchRequested.bind(this)}
                onSuggestionsClearRequested={this.onSuggestionsClearRequested.bind(this)}
                getSuggestionValue={this.getSuggestionValue.bind(this)}
                renderSuggestion={this.renderSuggestion.bind(this)}
            />  
        );
    }
}

export default Search;