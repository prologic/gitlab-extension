import React, {Component} from 'react';
import 'bootstrap/dist/css/bootstrap.css';

class Settings extends Component {
    constructor(props) {
        super(props);
    }

    render() {
        return (
            <div className="row">
                <ul className="list-group list-group-horizontal-md w-100 cornered">
                    <li className="list-group-item flex-fill">Cras justo odio</li>
                    <li className="list-group-item flex-fill">Dapibus ac facilisis in</li>
                    <li className="list-group-item flex-fill">Morbi leo risus</li>
                </ul>
            </div>
        )
    }
}

export default Settings;