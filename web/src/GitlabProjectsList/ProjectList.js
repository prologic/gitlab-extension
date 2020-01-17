import React, {Component} from 'react';
import 'bootstrap/dist/css/bootstrap.css';
import GitlabProject from "../GitlabProject/GitlabProject";
import Websocket from "react-websocket";

let loc = window.location, ws;
if (loc.protocol === "https:") {
    ws = "wss:";
} else {
    ws = "ws:";
}
ws += "//" + loc.host;
ws += loc.pathname + "ws";

const api_url = process.env.REACT_APP_API_URL ? process.env.REACT_APP_API_URL : "";
const ws_url = process.env.REACT_APP_WS_URL ? process.env.REACT_APP_WS_URL : ws;

class ProjectList extends Component {
    constructor(props) {
        super(props);
        this.state = {
            projects: [],
            isLoading: true,
            error: null
        };
    }

    componentDidMount() {
        this.callApi()
            .then(response => {
                this.setState({projects: response.projects, isLoading: false});
            })
            .catch(e => {
                this.setState({projects: null, isLoading: false, error: e})
            })
    }

    callApi = async () => {
        const response = await fetch(api_url + "/projects");
        const body = await response.json();
        if (response.status !== 200) throw Error(body.message);
        return body;
    };

    handlePush(data) {
        let gitlab_push = JSON.parse(data);
        this.setState(state => {
            let projects = [];
            Object.assign(projects, state.projects);
            for (let i = 0; i < projects.length; i++) {
                if (projects[i]["id"] === gitlab_push["project"]["id"]) {
                    let pipelines = projects[i]["pipelines"];
                    let updated = false;
                    if (pipelines) {
                        for (let j = 0; j < pipelines.length; j++) {
                            if (pipelines[j]["id"] === gitlab_push["object_attributes"]["id"]) {
                                pipelines[j]["status"] = gitlab_push["object_attributes"]["status"];
                                updated = true;
                                break;
                            }
                        }
                    }
                    if (!updated) {
                        let new_pipeline = {
                            id: gitlab_push["object_attributes"]["id"],
                            sha: gitlab_push["object_attributes"]["sha"],
                            branch: gitlab_push["object_attributes"]["ref"],
                            status: gitlab_push["object_attributes"]["status"],
                            web_url: gitlab_push["commit"]["url"],
                            commit: {
                                author: gitlab_push["commit"]["author"]["name"],
                                created_at: gitlab_push["commit"]["timestamp"],
                                title: gitlab_push["commit"]["message"]
                            }
                        };
                        pipelines.push(new_pipeline);
                        pipelines.sort((a, b) => b["id"] - a["id"]);
                        projects[i]["pipelines"] = pipelines
                    }
                    break;
                }
            }
            return {projects: projects}
        })
    }

    render() {
        const {projects, isLoading, error} = this.state;
        if (isLoading)
            return (
                <div className="vh-100 text-center vertical-20">
                    <div className="spinner-border spinner-large text-light" role="status">
                        <span className="sr-only">Loading...</span>
                    </div>
                </div>
            );
        if (error) return <div>Error: {error.message}</div>;
        else {
            const projectsMap = projects.map((item, key) =>
                <GitlabProject key={key} data={item}/>
            );
            return (
                <div>
                    {/*<Settings />*/}
                    {/*<div className="row">*/}
                    <div className="list-group shrinked">
                        <Websocket url={ws_url}
                                   onMessage={this.handlePush.bind(this)}/>
                        {projectsMap}
                    </div>
                    {/*</div>*/}
                </div>
            )
        }
    }
}

export default ProjectList;