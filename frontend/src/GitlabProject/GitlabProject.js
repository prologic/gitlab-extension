import React, {Component} from 'react';
import 'bootstrap/dist/css/bootstrap.css';
import GitlabPipeline from "../GitlabPipeline/GitlabPipeline"
import Moment from "react-moment";

class GitlabProject extends Component {
    constructor(props) {
        super(props);
        this.state = {
            data: props.data
        };
    }

    static createPipelines(pipelines) {
        if (!pipelines) return [];
        pipelines.sort((a, b) => a["id"] > b["id"]);
        return pipelines.map(item => <GitlabPipeline key={item["id"]} data={item}/>)
    }

    render() {
        const last_activity = this.state.data["last_activity"];
        const pipelines = GitlabProject.createPipelines(this.state.data["pipelines"]);
        return (
            <div>
                <div className="card card-small project-bg">
                    <div className="card-header padding-small">
                        <div className="d-flex flex-sm-row">
                            <div className="pl-1 ml-2">
                                <strong>{this.state.data.name}</strong>
                            </div>
                            <div className="flex-fill ">
                                <div className="float-right width-180" data-toggle="tooltip" title={last_activity}>
                                    Modified&nbsp;
                                    <Moment toNow>
                                        {last_activity}
                                    </Moment>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="ml-3 mr-3">
                    <ul className="list-group list-group-flush">
                        {pipelines}
                    </ul>
                </div>
            </div>
        )
    }
}

export default GitlabProject;
