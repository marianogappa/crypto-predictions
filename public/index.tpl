<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
    <style>
        /* roboto-regular - latin */
        @font-face {
        font-family: 'Roboto';
        font-style: normal;
        font-weight: 400;
        src: local(''),
            url('roboto-v30-latin-regular.woff2') format('woff2'), /* Super Modern Browsers */
            url('roboto-v30-latin-regular.woff') format('woff'), /* Modern Browsers */
        }

        .chartWithMarkerOverlay {
            position: relative;
            margin-top: -60px;
            /* border: 1px solid red; */
        }

        .footnote {
            /* border: 1px solid red; */
            color: white;
            font-family: Roboto, sans-serif;
            font-size: 12px;
            text-align: center;
        }

        #line-chart-marker {
            /* border: 1px solid yellow; */
            height: 408px;
        }

        .topSection {
            margin: 40px;
            display: flex;
            flex-direction: row;
            flex-wrap: wrap;
            margin-left: 80px;
            margin-right: 80px;
        }

        .topLeft,
        .topRight {
            /* border: 1px solid red; */
            display: flex;
            flex-direction: column;
            flex-basis: 100%;
            flex: 1;
        }


        .overlay-marker {
            font-size: 10px;
            font-family: 'Courier New', Courier, monospace;
            background-color: black;
            color: white;
            position: absolute;
            padding: 3px;
            opacity: 0.5;
        }

        html,
        body {
            width: 1200px;
            height: 675px;
            margin: 0;
            padding: 0;
            background-color: black;
            /* border: 1px solid white; */
            background-image: url('background.jpg');
        }

        .postAuthor {
            display: inline;
            position: absolute;
            left: 80px;
            font-size: 12px;
            color: white;
            top: 221px;
            font-family: Roboto, sans-serif;
            font-weight: bolder;
            width: 220px;
            text-align: center;
            text-shadow: 2px 2px 2px #000;
            /* border: 1px solid red; */

        }

        .bubble {
            font-family: sans-serif;
            font-size: 18px;
            line-height: 24px;
            width: 300px;
            border-radius: 40px;
            text-align: center;
            color: #000;
            padding: 14px;
            display: inline;
            position: absolute;
            left: 293px;
            background-color: #FFF;
            opacity: 0.8;
            top: 32px;
        }

        .bubble-bottom-left:before {
            content: "";
            width: 0px;
            height: 0px;
            position: absolute;
            border-left: 24px solid #FFF;
            border-right: 12px solid transparent;
            border-top: 12px solid #FFF;
            border-bottom: 20px solid transparent;
            left: 32px;
            bottom: -24px;
        }

        .happyface {
            border-radius: 50%;
            overflow: hidden;
            border: 5px solid #FFF;
            width: 13rem;
            height: 13rem;
            background-size: cover;
        }

        .predictionResult {
            font-size: 50px;
            color: white;
            font-family: Roboto, sans-serif;
            font-weight: bolder;
            text-align: right;
            margin-top: 49px;
        }

        .overlay {
            position: absolute;
            height: 0;
            width: 0;
        }

        #overlayUpperGreenBox {
            border-bottom: 1px dashed #00c30f;
            background-color: rgba(0, 195, 15, .2);
        }

        #overlayUpperGreenYellowBox {
            border-bottom: 1px dashed yellow;
            background-color: rgba(0, 195, 15, .2);
        }

        #overlayUpperRedBox {
            border-bottom: 1px dashed red;
            background-color: rgba(255, 0, 0, .2);
        }

        #overlayUpperRedYellowBox {
            border-bottom: 1px dashed red;
            background-color: rgba(255, 0, 0, .2);
        }

        #overlayLowerGreenBox {
            border-top: 1px dashed #00c30f;
            background-color: rgba(0, 195, 15, .2);
        }

        #overlayLowerGreenYellowBox {
            border-top: 1px dashed yellow;
            background-color: rgba(0, 195, 15, .2);
        }

        #overlayLowerRedBox {
            border-top: 1px dashed red;
            background-color: rgba(255, 0, 0, .2);
        }

        #overlayLowerRedYellowBox {
            border-top: 1px dashed red;
            background-color: rgba(255, 0, 0, .2);
        }

        #overlayUpperGreenYellowBoxLabel,
        #overlayUpperGreenBoxLabel,
        #overlayUpperRedYellowBoxLabel,
        #overlayUpperRedBoxLabel,
        #overlayLowerGreenYellowBoxLabel,
        #overlayLowerGreenBoxLabel,
        #overlayLowerRedYellowBoxLabel,
        #overlayLowerRedBoxLabel {
            color: white;
            font-family: 'Roboto', sans-serif;
            font-size: 11px;
        }

        #overlayPostedAtLine {
            border: 1px solid white;
        }

        #overlayEndedAtLine {
            border: 1px solid white;
        }
    </style>
    <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
    <!-- <script type="text/javascript" src="data.js"></script> -->
    <script type="text/javascript">
        const prediction = {{.Prediction }}
        const account = {{.Account }}
    </script>
    <script type="text/javascript">
        const events = [
            {
                "eventType": "entered",
                "target": 1,
                "price": 30860.77,
                "at": "2021-06-22T15:22:00Z",
                "takeProfitRatio": 0
            },
            {
                "eventType": "invalidated",
                "price": 34261.01,
                "at": "2021-06-24T15:22:00Z",
                "takeProfitRatio": 0.110180011711
            }
        ]
        function compileDataset() {
            const rawDataset = prediction.summary.candlestickMap[prediction.summary.coin]
            const dataset = []
            rawDataset.forEach(c => dataset.push([new Date(c.t * 1000), c.l, c.o, c.c, c.h]))
            return dataset
        }
        const dataset = compileDataset()

    </script>
    <script type="text/javascript">
        google.charts.load('current', { 'packages': ['corechart'] });
        google.charts.setOnLoadCallback(drawChart);

        function drawChart() {
            var data = new google.visualization.DataTable();
            data.addColumn('datetime')
            data.addColumn('number')
            data.addColumn('number')
            data.addColumn('number')
            data.addColumn('number')
            dataset.forEach(datapoint => data.addRow(datapoint))


            var options = {
                legend: 'none',
                candlestick: {
                    fallingColor: { strokeWidth: 0, fill: '#a52714' }, // red
                    risingColor: { strokeWidth: 0, fill: '#0f9d58' },   // green
                    // bar: { groupWidth: '100%' },
                },
                colors: ['#CCCCCC'],
                backgroundColor: { fill: 'transparent' },
                vAxis: {
                    gridlines: {
                        color: 'transparent'
                    },
                    textStyle: { color: '#FFF', fontSize: 9 },
                    viewWindowMode: 'explicit',
                    // viewWindow: {
                    //   max:3000,
                    //   min:500
                    // }
                },
                hAxis: {
                    gridlines: {
                        color: 'transparent'
                    },
                    textStyle: { color: '#FFF', fontSize: 9 }
                }
            };

            let upperGreenLineAt = null
            let lowerGreenLineAt = null
            let upperRedLineAt = null
            let lowerRedLineAt = null
            let postedAtLineAt = null
            let endedAtLineAt = null
            if (prediction.summary.goal && (prediction.summary.operator.startsWith('>'))) {
                options.vAxis.maxValue = prediction.summary.goal * 1.01
                upperGreenLineAt = [prediction.summary.goal, prediction.summary.goalWithError]
            }

            if (prediction.summary.goal && (prediction.summary.operator.startsWith('<'))) {
                options.vAxis.minValue = prediction.summary.goal * 0.99
                lowerGreenLineAt = [prediction.summary.goal, prediction.summary.goalWithError]
            }

            if (prediction.summary.rangeLow && prediction.summary.rangeHigh) {
                options.vAxis.minValue = prediction.summary.rangeLow * 0.99
                options.vAxis.maxValue = prediction.summary.rangeHigh * 1.01
                lowerRedLineAt = [prediction.summary.rangeLow, prediction.summary.rangeLowWithError]
                upperRedLineAt = [prediction.summary.rangeLow, prediction.summary.rangeHighWithError]
            }

            if (prediction.summary.willReach && prediction.summary.beforeItReaches) {
                options.vAxis.minValue = Math.min(prediction.summary.willReach, prediction.summary.beforeItReaches) * 0.99
                options.vAxis.maxValue = Math.max(prediction.summary.willReach, prediction.summary.beforeItReaches) * 1.01
                // TODO
                greenLineAt = prediction.summary.willReach
                redLineAt = prediction.summary.beforeItReaches
            }

            const firstCoinCandles = prediction.summary.candlestickMap[Object.keys(prediction.summary.candlestickMap)[0]]

            const deadlineTs = prediction.summary.endedAt ? new Date(prediction.summary.endedAt).getTime() / 1000 : null
            if (deadlineTs && prediction.state.status === 'FINISHED') {
                endedAtLineAt = new Date(prediction.summary.endedAt)
            }

            const postedAtTs = new Date(prediction.postedAt).getTime() / 1000
            if (postedAtTs && postedAtTs >= firstCoinCandles[0].t && postedAtTs <= firstCoinCandles[firstCoinCandles.length - 1].t) {
                postedAtLineAt = new Date(prediction.postedAt)
            }


            function placeMarkers(dataTable) {
                var cli = this.getChartLayoutInterface();
                var chartArea = cli.getChartAreaBoundingBox();

                let left = `${chartArea.left}px`
                let width = `${chartArea.width}px`
                if (postedAtLineAt) {
                    left = `${cli.getXLocation(postedAtLineAt)}px`
                    width = `${chartArea.width - (cli.getXLocation(postedAtLineAt) - chartArea.left)}px`
                }

                // Show upper green box, line and label
                if (upperGreenLineAt) {
                    const y1 = cli.getYLocation(upperGreenLineAt[0])

                    const e = document.querySelector('#overlayUpperGreenBox')
                    e.style.left = left
                    e.style.top = `${chartArea.top}px`
                    e.style.height = `${y1 - chartArea.top}px`
                    e.style.width = width

                    const label = document.querySelector('#overlayUpperGreenBoxLabel')
                    label.innerText = `Goal`
                    const endedAtPadding = endedAtLineAt ? 15 : 0
                    label.style.left = `${chartArea.left + chartArea.width + 5 + endedAtPadding}px`
                    label.style.height = 'auto'
                    label.style.width = 'auto'
                    label.style.top = `${y1 - (label.clientHeight / 2) + 1}px`;

                    // Show yellow box, line and label
                    const y2 = cli.getYLocation(upperGreenLineAt[1])
                    if (Math.abs(y1 - y2) > 15) {
                        const e = document.querySelector('#overlayUpperGreenYellowBox')
                        e.style.left = left
                        e.style.top = `${y1}px`
                        e.style.height = `${y2 - y1}px`
                        e.style.width = width

                        const label = document.querySelector('#overlayUpperGreenYellowBoxLabel')
                        label.innerText = `${prediction.summary.errorMarginRatio * 100}% error margin`
                        const endedAtPadding = endedAtLineAt ? 15 : 0
                        label.style.left = `${chartArea.left + chartArea.width + 5 + endedAtPadding}px`
                        label.style.height = 'auto'
                        label.style.width = 'auto'
                        label.style.top = `${y2 - (label.clientHeight / 2) + 1}px`;
                    }
                }

                // Show lower green box, line and label
                if (lowerGreenLineAt) {
                    const y1 = cli.getYLocation(lowerGreenLineAt[0])

                    const e = document.querySelector('#overlayLowerGreenBox')
                    e.style.left = left
                    e.style.top = `${y1}px`
                    e.style.height = `${chartArea.top + chartArea.height - y1}px`
                    e.style.width = width

                    const label = document.querySelector('#overlayLowerGreenBoxLabel')
                    label.innerText = `Goal`
                    const endedAtPadding = endedAtLineAt ? 15 : 0
                    label.style.left = `${chartArea.left + chartArea.width + 5 + endedAtPadding}px`
                    label.style.height = 'auto'
                    label.style.width = 'auto'
                    label.style.top = `${y1 - (label.clientHeight / 2) + 1}px`;

                    // Show yellow box, line and label
                    const y2 = cli.getYLocation(lowerGreenLineAt[1])
                    if (Math.abs(y1 - y2) > 15) {
                        const e = document.querySelector('#overlayLowerGreenYellowBox')
                        e.style.left = left
                        e.style.top = `${y2}px`
                        e.style.height = `${y1 - y2}px`
                        e.style.width = width

                        const label = document.querySelector('#overlayLowerGreenYellowBoxLabel')
                        label.innerText = `${prediction.summary.errorMarginRatio * 100}% error margin`
                        const endedAtPadding = endedAtLineAt ? 15 : 0
                        label.style.left = `${chartArea.left + chartArea.width + 5 + endedAtPadding}px`
                        label.style.height = 'auto'
                        label.style.width = 'auto'
                        label.style.top = `${y2 - (label.clientHeight / 2) + 1}px`;
                    }
                }

                // Show upper red box, line and label
                if (upperRedLineAt) {
                    const y1 = cli.getYLocation(upperRedLineAt[0])

                    const e = document.querySelector('#overlayUpperRedBox')
                    e.style.left = left
                    e.style.top = `${chartArea.top}px`
                    e.style.height = `${y1 - chartArea.top}px`
                    e.style.width = width

                    const label = document.querySelector('#overlayUpperRedBoxLabel')
                    label.innerText = `Goal`
                    const endedAtPadding = endedAtLineAt ? 15 : 0
                    label.style.left = `${chartArea.left + chartArea.width + 5 + endedAtPadding}px`
                    label.style.height = 'auto'
                    label.style.width = 'auto'
                    label.style.top = `${y1 - (label.clientHeight / 2) + 1}px`;

                    // Show yellow box, line and label
                    const y2 = cli.getYLocation(upperRedLineAt[1])
                    if (Math.abs(y1 - y2) > 15) {
                        const e = document.querySelector('#overlayUpperRedYellowBox')
                        e.style.left = left
                        e.style.top = `${y1}px`
                        e.style.height = `${y2 - y1}px`
                        e.style.width = width

                        const label = document.querySelector('#overlayUpperRedYellowBoxLabel')
                        label.innerText = `${prediction.summary.errorMarginRatio * 100}% error margin`
                        const endedAtPadding = endedAtLineAt ? 15 : 0
                        label.style.left = `${chartArea.left + chartArea.width + 5 + endedAtPadding}px`
                        label.style.height = 'auto'
                        label.style.width = 'auto'
                        label.style.top = `${y2 - (label.clientHeight / 2) + 1}px`;
                    }
                }

                // Show lower Red box, line and label
                if (lowerRedLineAt) {
                    const y1 = cli.getYLocation(lowerRedLineAt[0])

                    const e = document.querySelector('#overlayLowerRedBox')
                    e.style.left = left
                    e.style.top = `${y1}px`
                    e.style.height = `${chartArea.top + chartArea.height - y1}px`
                    e.style.width = width

                    const label = document.querySelector('#overlayLowerRedBoxLabel')
                    label.innerText = `Goal`
                    const endedAtPadding = endedAtLineAt ? 15 : 0
                    label.style.left = `${chartArea.left + chartArea.width + 5 + endedAtPadding}px`
                    label.style.height = 'auto'
                    label.style.width = 'auto'
                    label.style.top = `${y1 - (label.clientHeight / 2) + 1}px`;

                    // Show yellow box, line and label
                    const y2 = cli.getYLocation(lowerRedLineAt[1])
                    if (Math.abs(y1 - y2) > 15) {
                        const e = document.querySelector('#overlayLowerRedYellowBox')
                        e.style.left = left
                        e.style.top = `${y2}px`
                        e.style.height = `${y1 - y2}px`
                        e.style.width = width

                        const label = document.querySelector('#overlayLowerRedYellowBoxLabel')
                        label.innerText = `${prediction.summary.errorMarginRatio * 100}% error margin`
                        const endedAtPadding = endedAtLineAt ? 15 : 0
                        label.style.left = `${chartArea.left + chartArea.width + 5 + endedAtPadding}px`
                        label.style.height = 'auto'
                        label.style.width = 'auto'
                        label.style.top = `${y2 - (label.clientHeight / 2) + 1}px`;
                    }
                }


                // Show "finish line" white line with checkered flag emoji
                if (endedAtLineAt) {
                    const x = cli.getXLocation(endedAtLineAt)
                    const e = document.querySelector('#overlayEndedAtLine')
                    e.style.left = `${x}px`;
                    e.style.top = `${chartArea.top}px`
                    e.style.height = `${chartArea.height}px`

                    const label = document.querySelector('#overlayEndedAtLabel')
                    label.innerText = `🏁`
                    label.style.height = 'auto'
                    label.style.width = 'auto'
                    label.style.left = `${x - (label.clientWidth / 2) + 1}px`;
                    label.style.top = `${chartArea.top - label.clientHeight}px`
                }

                // Show "starting line" white line with eyes emoji
                if (postedAtLineAt) {
                    const x = cli.getXLocation(postedAtLineAt)
                    const e = document.querySelector('#overlayPostedAtLine')
                    e.style.left = `${x}px`;
                    e.style.top = `${chartArea.top}px`
                    e.style.height = `${chartArea.height}px`

                    const label = document.querySelector('#overlayPostedAtLineLabel')
                    label.innerText = `👀`
                    label.style.height = 'auto'
                    label.style.width = 'auto'
                    label.style.left = `${x - (label.clientWidth / 2) + 1}px`;
                    label.style.top = `${chartArea.top - label.clientHeight}px`
                }
            };

            var chart = new google.visualization.CandlestickChart(document.getElementById('line-chart-marker'));

            google.visualization.events.addListener(chart, 'ready',
                placeMarkers.bind(chart, data)
            );
            chart.draw(data, options);
        }
    </script>
</head>

<body>
    <div class="topSection">
        <div class="topLeft">
            <img class="happyface" src="" />
            <div class="bubble bubble-bottom-left"></div>
            <span class="postAuthor"></span>
        </div>
        <div class="topRight">
            <span class="predictionResult"></span>
        </div>
    </div>
    <div class="chartWithMarkerOverlay">

        <div id="line-chart-marker"></div>

        <div class="overlay" id="overlayUpperGreenBox"></div>
        <div class="overlay" id="overlayUpperGreenBoxLabel"></div>
        <div class="overlay" id="overlayUpperGreenYellowBox"></div>
        <div class="overlay" id="overlayUpperGreenYellowBoxLabel"></div>
        <div class="overlay" id="overlayLowerGreenBox"></div>
        <div class="overlay" id="overlayLowerGreenBoxLabel"></div>
        <div class="overlay" id="overlayLowerGreenYellowBox"></div>
        <div class="overlay" id="overlayLowerGreenYellowBoxLabel"></div>

        <div class="overlay" id="overlayUpperRedBox"></div>
        <div class="overlay" id="overlayUpperRedBoxLabel"></div>
        <div class="overlay" id="overlayUpperRedYellowBox"></div>
        <div class="overlay" id="overlayUpperRedYellowBoxLabel"></div>
        <div class="overlay" id="overlayLowerRedBox"></div>
        <div class="overlay" id="overlayLowerRedBoxLabel"></div>
        <div class="overlay" id="overlayLowerRedYellowBox"></div>
        <div class="overlay" id="overlayLowerRedYellowBoxLabel"></div>

        <div class="overlay" id="overlayPostedAtLine"></div>
        <div class="overlay" id="overlayPostedAtLineLabel"></div>

        <div class="overlay" id="overlayEndedAtLine"></div>
        <div class="overlay" id="overlayEndedAtLabel"></div>

        <!-- <div class="overlay-marker-1 overlay-marker">
            <span>▲ ENTRY</span>
        </div>
        <div class="overlay-marker-2 overlay-marker">
            <span>▼ EXIT</span>
        </div> -->

    </div>
    <div class="footnote">* Goal may be reached at latest incomplete
        candlestick
        and not shown. Tweets don't store timezone information, which makes "end of day" and the like slightly
        innacurate.</div>
    <script type="text/javascript">
        document.querySelector('.happyface').setAttribute("src", account.thumbnails[account.thumbnails.length - 1]);
        document.querySelector('.bubble').innerText = prediction.predictionText;

        if (prediction.state.status !== 'FINISHED') {
            document.querySelector('.predictionResult').innerText = 'NOW TRACKING 👀';
        } else if (prediction.state.value === 'CORRECT') {
            document.querySelector('.predictionResult').innerText = 'CORRECT ✅';
        } else {
            document.querySelector('.predictionResult').innerText = 'INCORRECT ❌';
        }

        if (account.handle) {
            document.querySelector('.postAuthor').innerText = `@${account.handle}`;
        } else {
            document.querySelector('.postAuthor').innerText = account.name;
        }
    </script>

</body>

</html>
