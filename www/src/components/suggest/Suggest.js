import React, { Component, PropTypes } from 'react';
import './Suggest.css';

import Autosuggest from 'react-autosuggest';
import { ServerUrl } from '../../constants';

class Suggest extends Component {
    static propTypes = {
        suggestionSelected: PropTypes.func
    }

    state = {
        suggestions: [],
        value: ''
    }

    onChange(event, { newValue, method }) {
        this.setState({value: newValue});
    }
    
    renderSuggestion(suggestion) {
        return <div className="suggestion-item">{suggestion}</div>
    }

    getSuggestionValue(suggestion) {
        return suggestion;
    }

    onSuggestionsFetchRequested(text) {
        fetch(`${ServerUrl}/meta/autocomplete/${text.value}`)
            .then(response => response.json())
            .then(suggestions => this.setState({suggestions}))
            .catch(err => console.error(err));
    }

    onSuggestionsClearRequested() {
        this.setState({suggestions: []});
    }

    fireSuggestionSelected(event, { suggestionValue }) {
        this.props.suggestionSelected(suggestionValue);
    }

    handleSuggestKeyPressed({ key }) {
        if (key === 'Enter') {
            this.fireSuggestionSelected(null, {suggestionValue: this.state.value });
        }    
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
                onSuggestionSelected={this.fireSuggestionSelected.bind(this)}
                renderInputComponent={inputProps => (
                    <div>
                        <input onKeyPress={this.handleSuggestKeyPressed.bind(this)} {...inputProps} />
                    </div>
                )}
            />  
        );
    }
}

export default Suggest;