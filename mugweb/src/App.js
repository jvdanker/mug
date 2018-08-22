import React, {Component} from 'react';
import './App.css';

class App extends Component {

    constructor(props) {
        super(props);

        this.state = {
            urls: [
                '1',
                '2',
                '3',
                '4',
                '5'
            ],
            url: ''
        };

        this.handleChange = this.handleChange.bind(this);
        this.addUrl = this.addUrl.bind(this);
    }

    handleChange(event) {
        this.setState({url: event.target.value});
    }

    addUrl(event) {
        event.preventDefault();

        const urls = this.state.urls;
        urls.push(this.state.url);

        this.setState({
            urls: urls
        });
    }

    scanLink(url, index, event) {
        event.preventDefault();

        fetch("http://localhost:8080/scan")
            .then(res => console.log(res))
            .catch(error => console.error('Error:', error));
    }

    deleteLink(url, index, event) {
        event.preventDefault();

        var urls = this.state.urls;
        urls.splice(index, 1);
        this.setState({
            urls: urls
        });
    }

    render() {
        const listItems = this.state.urls.map((url, index) =>
            <li key={index}>
                <div>
                    {url}
                </div>
                <div>
                    <a href="scan-link.html" onClick={this.scanLink.bind(this, url, index)}>scan</a>
                    <a href="delete" onClick={this.deleteLink.bind(this, url, index)}>delete</a>
                </div>
            </li>
        );

        return (
            <div className="App">
                <form onSubmit={this.addUrl}>
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
