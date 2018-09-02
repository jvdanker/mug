import React, {Component} from 'react';
import styled from 'styled-components';
import './App.css';

const ImageContainer = styled.div`
  width: 100px;
  height: 100px;
  overflow: hidden;
`;

const Image = styled.img`
  width: 100px;
`;

class App extends Component {

    constructor(props) {
        super(props);

        this.state = {
            urls: [],
            url: 'https://www.govt.nz/'
        };

        this.timerId = -1;

        this.handleChange = this.handleChange.bind(this);
        this.addUrl = this.addUrl.bind(this);
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
    }

    componentWillUnmount() {
        if (this.timerId !== -1) {
            clearInterval(this.timerId);
        }
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

                console.log(this.timerId);
                this.timerId = setTimeout(
                    () => {
                        console.log('tick');
                        this.getReference(response.id);
                    }, 5000
                );

            })
            .catch(error => console.error('Error:', error));
    }

    scanAll(event) {
        event.preventDefault();
        var state = this.state;

        fetch("http://localhost:8080/scan",{
            method: 'POST',
            body: JSON.stringify({type: 'current'}),
            headers: {
                'Content-Type': 'application/json'
            }
        }).then(res => res.json())
            .then(response => {
                console.log(response);

                this.timerId = setInterval(
                    () => {
                        console.log('tick', response.ids.length, response.ids);

                        if (response.ids.length === 0) {
                            clearInterval(this.timerId);
                            this.timerId = -1;
                        } else {
                            var id = response.ids.splice(0, 1)[0];

                            fetch("http://localhost:8080/screenshot/scan/" + id)
                                .then(res => res.json())
                                .then(res => {
                                    console.log(res);

                                    var urls = state.urls;
                                    var index = urls.findIndex(e => {
                                        return e.id === id;
                                    });

                                    if (index > -1) {
                                        urls[index].current = res.data;
                                        this.setState({
                                            urls: urls
                                        });
                                    } else {
                                        console.error("Index not found: ", id);
                                    }
                                }).catch(error => console.error('Error:', error));
                        }
                    }, 5000);
            }).catch(error => console.error('Error:', error));
    }

    scanLink(item, event) {
        event.preventDefault();

        fetch("http://localhost:8080/url/scan/" + item.id)
            .then(res => res.json())
            .then(response => {
                this.timerId = setTimeout(
                    () => {
                        this.getScan(item.id);
                    }, 5000
                );

            })
            .catch(error => console.error('Error:', error));
    }

    getScan(id) {
        fetch("http://localhost:8080/screenshot/scan/" + id)
            .then(res => res.json())
            .then(res => {
                console.log(res);

                var urls = this.state.urls;
                var index = urls.findIndex(e => {
                    return e.id === id;
                });

                if (index > -1) {
                    urls[index].current = res.data;
                    this.setState({
                        urls: urls
                    });
                }

                clearInterval(this.timerId);
                this.timerId = -1;
            }).catch(error => {
                console.error('Error:', error);

                this.timerId = setTimeout(
                    () => {
                        console.log('tick');
                        this.getScan(id);
                    }, 5000);
            });
    }

    initLink(item, event) {
        event.preventDefault();

        fetch("http://localhost:8080/init/" + item.id)
            .then(res => console.log(res))
            .catch(error => console.error('Error:', error));
    }

    diffLink(item, event) {
        event.preventDefault();

        fetch("http://localhost:8080/pdiff/" + item.id)
            .then(res => res.json())
            .then(res => {
                console.log(res);

                var urls = this.state.urls;
                var index = urls.findIndex(e => {
                    return e.id === item.id;
                });

                if (index > -1) {
                    urls[index].output = res.output;
                    this.setState({
                        urls: urls
                    });
                }
            })
            .catch(error => console.error('Error:', error));
    }

    deleteLink(item, event) {
        event.preventDefault();

        var urls = this.state.urls;
        var index = urls.findIndex(e => {
            return e.id === item.id;
        });

        if (index > -1) {
            fetch("http://localhost:8080/url/delete/" + item.id)
                .then(res => console.log(res))
                .catch(error => console.error('Error:', error));


            urls.splice(index, 1);
            this.setState({
                urls: urls
            });
        }
    }

    getReference(id, event) {
        if (event) {
            event.preventDefault();
        }

        fetch("http://localhost:8080/screenshot/reference/get/" + id)
            .then(res => res.json())
            .then(res => {
                console.log(res);

                var urls = this.state.urls;
                var index = urls.findIndex(e => {
                    return e.id === id;
                });

                if (index > -1) {
                    urls[index].reference = res.data;
                    this.setState({
                        urls: urls
                    });
                }

                clearInterval(this.timerId);
                this.timerId = -1;
            }).catch(error => {
                console.error('Error:', error);

                this.timerId = setTimeout(
                    () => {
                        console.log('tick');
                        this.getReference(id);
                    }, 5000);
            });
    }

    render() {
        const listItems = this.state.urls.map((item, index) =>
            <li key={index}>
                <div>
                    <ImageContainer>
                        <Image src={item.reference} />
                    </ImageContainer>
                    <ImageContainer>
                        <Image src={item.current} />
                    </ImageContainer>
                    <div>
                        <pre>
                            {item.output}
                        </pre>
                    </div>
                    <div>
                        {item.url}
                    </div>
                    <div>
                        <a href="scan-link.html" onClick={this.scanLink.bind(this, item)}>scan</a>&nbsp;
                        <a href="delete" onClick={this.deleteLink.bind(this, item)}>delete</a>&nbsp;
                        <a href="init" onClick={this.initLink.bind(this, item)}>init</a>&nbsp;
                        <a href="diff" onClick={this.diffLink.bind(this, item)}>diff</a>&nbsp;
                    </div>
                </div>
            </li>
        );

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
                </form>

                <ul>{listItems}</ul>
            </div>
        );
    }
}

export default App;
