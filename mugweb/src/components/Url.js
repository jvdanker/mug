import  React, {Component} from 'react';
import ImageContainer from "./ImageContainer";
import styled from "styled-components";

const StyledContainer = styled.div`
    border: 1px solid black;
    display: flex;
    align-items: center;
    padding: 10px;
`;

const StyledUrl = styled.div`
    flex: 1;
    text-align: left;
`;

class Url extends Component {

    scanLink(event) {
        event.preventDefault();
        this.props.onScan(this.props.item);
    }

    deleteLink(event) {
        event.preventDefault();
        this.props.onDelete(this.props.item);
    }

    initLink(event) {
        event.preventDefault();
        this.props.onInit(this.props.item);
    }

    diffLink(event) {
        event.preventDefault();
        this.props.onDiff(this.props.item);
    }

    render() {
        return (
            <StyledContainer>
                <ImageContainer image={this.props.item.reference} />
                <ImageContainer image={this.props.item.current} />
                <StyledUrl>
                    {this.props.item.url}
                </StyledUrl>
                <div>
                    <pre>
                        {this.props.item.results}
                    </pre>
                </div>
                <div>
                    {this.props.item.status}
                </div>
                <div>
                    <a href="scan-link.html" onClick={this.scanLink.bind(this)}>scan</a>&nbsp;
                    <a href="delete" onClick={this.deleteLink.bind(this)}>delete</a>&nbsp;
                    <a href="init" onClick={this.initLink.bind(this)}>init</a>&nbsp;
                    <a href="diff" onClick={this.diffLink.bind(this)}>diff</a>&nbsp;
                </div>
            </StyledContainer>
        );
    }
}

export default Url;
