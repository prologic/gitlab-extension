import React, {Component} from 'react';
import 'bootstrap/dist/css/bootstrap.css';

class GitlabPipeline extends Component {
    constructor(props) {
        super(props);
        this.state = {
            data: this.props.data
        };
        this.statusClass = this.statusClass.bind(this)
    }

    static makeBadge(kind) {
        return "badge width-120 " + kind;
    }

    statusClass() {
        switch (this.state.data["status"]) {
            case "running" :
                return GitlabPipeline.makeBadge("badge-primary");
            case "pending" :
                return GitlabPipeline.makeBadge("badge-info");
            case "success" :
                return GitlabPipeline.makeBadge("badge-success");
            case "failed" :
                return GitlabPipeline.makeBadge("badge-danger");
            case "canceled" :
                return GitlabPipeline.makeBadge("badge-warning");
            case "skipped" :
                return GitlabPipeline.makeBadge("badge-secondary");
            default:
                return ""
        }
    }

    render() {
        const pipeline = this.state.data;
        const commit = pipeline["commit"];
        return (
            <li className="list-group-item padding-small">
                <div className="d-flex flex-sm-row">
                    <div className="width-200">{pipeline["branch"]}</div>
                    <div className="width-120">
                        <a href={pipeline["web_url"]}>{pipeline["id"]}</a>
                    </div>
                    <div>
                        {commit["author"] + " : " + commit["title"]}
                    </div>
                    <div className="p-0 flex-fill">
                        <div className="p-0 ml-4 float-right">
                            <span className={this.statusClass()}>Status: {pipeline["status"]}</span>
                        </div>
                        <div className="p-2 float-right">
                            {pipeline["duration"]}
                        </div>
                    </div>
                </div>
            </li>
        )
    }
}

export default GitlabPipeline;