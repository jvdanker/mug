import  React, {Component} from 'react';
import Url from "./Url";
import styled from 'styled-components';

const StyledContainer = styled.div`
  // width: 100px;
  // height: 100px;
  // overflow: hidden;
  
  li {
    // border: 1px solid black;
    list-style: none;
  }
`;

class Urls extends Component {

    render() {
        const listItems = this.props.urls.map((item, index) =>
            <li key={index}>
                <Url
                    item={item}
                    onScan={this.props.onScan}
                    onDelete={this.props.onDelete}
                    onInit={this.props.onInit}
                    onDiff={this.props.onDiff}
                />
            </li>
        );

        return (
            <StyledContainer>
                <ul>{listItems}</ul>
            </StyledContainer>
        );
    }
}

export default Urls;
