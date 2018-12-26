let stopCode = null;
let timer = null;

$(() => {
    let createNode = (element) => document.createElement(element);

    let append = (parent, el) => parent.appendChild(el);

    let load = (code) => {
        stopCode = code;
        fetch("http://localhost:5555/api/v1/stop/" + code)
            .then((resp) => resp.json())
            .then((data) => {
                $("#name").innerHTML = data.name;
                let stops = data.stops;
                return stops.map(function (stop) {
                    let tr = createNode('tr');
                    let line = createNode('td');
                    line.innerHTML = stop.line
                    let dest = createNode('td');
                    dest.innerHTML = stop.dest
                    let time = createNode('td');
                    time.innerHTML = stop.time
                    let eta = createNode('td');
                    eta.innerHTML = stop.eta
                    append(tr, line);
                    append(tr, dest);
                    append(tr, time);
                    append(tr, eta);
                    $("td").remove();
                    $("#stops").append(tr);
                })
            })
            .catch((error) => console.log(error));
    }

    $("#form_code").submit((event) => {
        load($("#code").val());
        return false;
    });

    $("#time_form").submit((event) => {
        if (stopCode != null) {
            window.clearInterval(timer);
            timer = window.setInterval(load, parseFloat($("#time").val()) * 1000, stopCode);
        }
        return false;
    });
});