import React, {Component} from 'react';
import './App.css';

class App extends Component {

    constructor(props) {
        super(props);

        this.state = {
            urls: [],
            url: ''
        };

        this.handleChange = this.handleChange.bind(this);
        this.add = this.add.bind(this);
    }

    handleChange(event) {
        this.setState({url: event.target.value});
    }

    add(event) {
        event.preventDefault();

        const urls = this.state.urls;
        urls.push(this.state.url);

        this.setState({
            urls: urls
        });
    }

    render() {
        const listItems = this.state.urls.map((url, index) =>
            <li key={index}>
                {url}
            </li>
        );

        return (
            <div className="App">
                <form onSubmit={this.add}>
                    <fieldset>
                        <legend>Add URL</legend>

                        <label htmlFor="url">Enter an URL:</label>
                        <input type="url"
                               name="url"
                               placeholder="https://example.com"
                               pattern="(http(s?)://?).*"
                               size="20"
                               value={this.state.url}
                               onChange={this.handleChange}
                               required/>
                    </fieldset>

                    <button type="submit">Add URL</button>
                </form>

                <ul>{listItems}</ul>
            </div>
        );
    }
}

export default App;
