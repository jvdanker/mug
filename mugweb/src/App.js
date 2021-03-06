import React, {Component} from 'react';
import Urls from './components/Urls';
import './App.css';

class App extends Component {

    constructor(props) {
        super(props);

        this.state = {
            urls: [],
            url: 'https://www.govt.nz/',
            updateTimerId: -1
        };

        this.handleChange = this.handleChange.bind(this);
        this.addUrl = this.addUrl.bind(this);

        this.scanLink = this.scanLink.bind(this);
        this.deleteLink = this.deleteLink.bind(this);
        this.initLink = this.initLink.bind(this);
        this.diffLink = this.diffLink.bind(this);
    }

    componentDidMount() {
        fetch("http://localhost:8080/list")
            .then(res => res.json())
            .then(res => {
                // console.log('json', JSON.stringify(res));
                this.setState({
                    urls: res
                });
            })
            .catch(error => console.error('Error:', error));

        // this.startUpdateTimer();
    }

    componentWillUnmount() {
        if (this.state.updateTimerId !== -1) {
            clearInterval(this.state.updateTimerId);
            this.setState({
                updateTimerId: -1
            });
        }
    }

    toggleTimer() {
        var id = this.state.updateTimerId;

        if (id === -1) {
            this.startUpdateTimer()
        } else {
            clearInterval(id);
            this.setState({
                updateTimerId: -1
            });
        }
    }

    startUpdateTimer() {
        console.log("Started update timer...");
        var updateTimerId = setInterval(() => this.timer(), 5000);

        this.setState({
            updateTimerId: updateTimerId
        });
    }

    timer() {
        fetch("http://localhost:8080/updates")
            .then(res => res.json())
            .then(res => {
                if (res.Type === undefined) return;
                console.log('tick', res);

                var id = res.Id;
                var urls = this.state.urls;
                var index = urls.findIndex(e => {
                    return e.id === id;
                });

                if (index > -1) {
                    console.log("type = ", res.Type);
                    switch (res.Type) {
                        case 0: // updated reference
                            urls[index].reference = res.Data.reference;
                            break;
                        case 1: // updated current
                            urls[index].current = res.Data.current;
                            break;
                        case 2: // updated diff
                            urls[index].results = res.Data.results;
                            urls[index].status = res.Data.status;
                            break;
                        default:
                            console.error("Unknown type = ", res.Type);
                            break;
                    }

                    this.setState({
                        urls: urls
                    });
                }
            });
    }

    handleChange(event) {
        this.setState({url: event.target.value});
    }

    addUrl(event) {
        event.preventDefault();

        const urls = this.state.urls;

        fetch("http://localhost:8080/url/add", {
            method: 'POST',
            body: JSON.stringify({url: this.state.url}),
            headers: {
                'Content-Type': 'application/json'
            }
        }).then(res => res.json())
            .then(response => {
                console.log('Success:', response);

                urls.push({
                    id: response.id,
                    url: this.state.url
                });

                this.setState({
                    urls: urls
                });
            })
            .catch(error => console.error('Error:', error));
    }

    scanAll(event) {
        event.preventDefault();

        fetch("http://localhost:8080/scan",{
            method: 'POST',
            body: JSON.stringify({type: 'current'}),
            headers: {
                'Content-Type': 'application/json'
            }
        }).then(res => res.json())
            .then(response => console.log(response))
            .catch(error => console.error('Error:', error));
    }

    scanLink(item) {
        fetch("http://localhost:8080/url/scan/" + item.id)
            .then(res => res.json())
            .then(response => console.log(response))
            .catch(error => console.error('Error:', error));
    }

    deleteLink(item) {
        var urls = this.state.urls;
        var index = urls.findIndex(e => {
            return e.id === item.id;
        });

        if (index > -1) {
            fetch("http://localhost:8080/url/" + item.id, {
                method: 'DELETE'
            }).then(res => console.log(res))
                .catch(error => console.error('Error:', error));


            urls.splice(index, 1);
            this.setState({
                urls: urls
            });
        }
    }

    initLink(item) {
        fetch("http://localhost:8080/init/" + item.id)
            .then(res => console.log(res))
            .catch(error => console.error('Error:', error));
    }

    diffLink(item) {
        fetch("http://localhost:8080/pdiff/" + item.id)
            .then(res => res.json())
            .then(res => {
                console.log(res);

                var urls = this.state.urls;
                var index = urls.findIndex(e => {
                    return e.id === item.id;
                });

                if (index > -1) {
                    urls[index].results = res.output;
                    this.setState({
                        urls: urls
                    });
                }
            })
            .catch(error => console.error('Error:', error));
    }

    render() {
        return (
            <div className="App">
                <form>
                    <fieldset>
                        <legend>Add URL</legend>

                        <label htmlFor="url">Enter a URL:</label>
                        <input type="url"
                               name="url"
                               placeholder="https://example.com"
                               pattern="(http(s?)://?).*"
                               size="20"
                               value={this.state.url}
                               onChange={this.handleChange}
                               required/>
                    </fieldset>

                    <button type="button" onClick={this.addUrl}>Add URL</button>
                    <button type="button" onClick={this.scanAll.bind(this)}>Scan all</button>
                    {this.state.updateTimerId !== -1 &&
                        <div>
                            Running
                            <button type="button" onClick={this.toggleTimer.bind(this)}>Stop updates</button>
                        </div>
                    }
                    {this.state.updateTimerId === -1 &&
                        <div>
                            Stopped
                            <button type="button" onClick={this.toggleTimer.bind(this)}>Start updates</button>
                        </div>
                    }
                </form>

                <Urls
                    urls={this.state.urls}
                    onScan={this.scanLink}
                    onDelete={this.deleteLink}
                    onInit={this.initLink}
                    onDiff={this.diffLink}
                />
            </div>
        );
    }
}

export default App;
