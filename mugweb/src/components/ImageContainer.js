import React, {Component} from 'react';
import styled from 'styled-components';

const Container = styled.div`
  width: 100px;
  height: 100px;
  overflow: hidden;
  display: flex;
`;

const Image = styled.img`
  width: 100px;
`;

class ImageContainer extends Component {

    render() {
        return (
            <Container>
                <Image src={this.props.image} />
            </Container>
        );
    }
}

export default ImageContainer;
