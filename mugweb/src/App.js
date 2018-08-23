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
        this.addUrl = this.addUrl.bind(this);
    }

    componentDidMount() {
        fetch("http://localhost:8080/list")
            .then(res => res.json())
            .then(res => {
                console.log('json', JSON.stringify(res));
                this.setState({
                    urls: res
                });
            })
            .catch(error => console.error('Error:', error));
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

    scanLink(item, event) {
        event.preventDefault();

        fetch("http://localhost:8080/scan/" + item.id)
            .then(res => console.log(res))
            .catch(error => console.error('Error:', error));
    }

    deleteLink(item, event) {
        event.preventDefault();

        var urls = this.state.urls;
        var index = urls.findIndex(e => {
            return e.id === item.id;
        });

        if (index > -1) {
            urls.splice(index, 1);
            this.setState({
                urls: urls
            });
        }
    }

    render() {
        const listItems = this.state.urls.map((item, index) =>
            <li key={index}>
                <div>
                    {item.url}
                </div>
                <div>
                    <a href="scan-link.html" onClick={this.scanLink.bind(this, item)}>scan</a>
                    <a href="delete" onClick={this.deleteLink.bind(this, item)}>delete</a>
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
