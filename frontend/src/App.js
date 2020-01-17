import React, {Component} from 'react';
import 'bootstrap/dist/css/bootstrap.css';
import ProjectList from "./GitlabProjectsList/ProjectList";


class App extends Component {
    render() {
        return (
            <div className="container-fluid bg-dark">
                <ProjectList />
            </div>
        )
    }
}

export default App;
