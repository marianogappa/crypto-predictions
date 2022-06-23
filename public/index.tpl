<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
    <script type="text/javascript">
        const prediction = {{.Prediction }}
        const account = {{.Account }}
    </script>
    <style>
        .chartWithMarkerOverlay {
            position: relative;
            margin-top: -60px;
            /* border: 1px solid red; */
        }

        .footnote {
            /* border: 1px solid red; */
            color: white;
            font-family: 'Roboto', sans-serif;
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
            flex: 1;
        }

        .topRight {
            margin-top: 49px;
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
            font-family: 'Roboto', sans-serif;
            font-weight: bolder;
            width: 220px;
            text-align: center;
            text-shadow: 2px 2px 2px #000;
            /* border: 1px solid red; */

        }

        .bubble {
            font-family: 'Roboto', sans-serif;
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

        .profileImage {
            border-radius: 50%;
            overflow: hidden;
            border: 5px solid #FFF;
            width: 13rem;
            height: 13rem;
            background-size: cover;
        }

        .filler {
            flex-grow: 1;
        }

        .predictionResult {
            font-size: 50px;
            color: white;
            font-family: 'Roboto', sans-serif;
        }

        #predictionResultImage {
            height: 50px;
            width: 50px;
            margin-left: 15px;
            margin-top: -2px;
        }

        .overlay {
            position: absolute;
            height: 0;
            width: 0;
        }

        .verticalLine {
            border: 1px solid white;
        }

        .postedAt,
        .endedAt {
            height: 20px;
            width: 20px;
        }

        .postedAt {
            content: url('eyes_1f440.png');
        }

        .endedAt {
            content: url('chequered-flag_1f3c1.png');
        }

        .greenLine {
            border-top: 1px dashed #00c30f;
        }

        .yellowLine {
            border-top: 1px dashed yellow;
        }

        .redLine {
            border-bottom: 1px dashed red;
        }

        .greenBox {
            background-color: rgba(0, 195, 15, .2);
        }

        .yellowBox {
            background-color: rgba(0, 195, 15, .2);
        }

        .redBox {
            background-color: rgba(255, 0, 0, .2);
        }

        .goalCaption {
            color: white;
            font-family: 'Roboto', sans-serif;
            font-size: 11px;
        }
    </style>
    <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
    <script type="text/javascript">
        function compileDataset() {
            const rawDataset = prediction.summary.candlestickMap[prediction.summary.coin]
            const dataset = []
            rawDataset.forEach(c => dataset.push([new Date(c.t * 1000), c.l, c.o, c.c, c.h]))
            return dataset
        }
        const dataset = compileDataset()

        google.charts.load('current', { 'packages': ['corechart'] });
        google.charts.setOnLoadCallback(drawChart);

        function drawChart() {
            // Configure the chart's candlestick data columns (i.e. date, & OHLC prices)
            var data = new google.visualization.DataTable();
            data.addColumn('datetime')
            data.addColumn('number')
            data.addColumn('number')
            data.addColumn('number')
            data.addColumn('number')
            dataset.forEach(datapoint => data.addRow(datapoint))

            // Google Chart options tweak the UI.
            var options = {
                legend: 'none',
                candlestick: {
                    fallingColor: { strokeWidth: 0, fill: '#a52714' }, // red
                    risingColor: { strokeWidth: 0, fill: '#0f9d58' },   // green
                },
                colors: ['#CCCCCC'],
                backgroundColor: { fill: 'transparent' },
                vAxis: {
                    gridlines: {
                        color: 'transparent'
                    },
                    textStyle: { color: '#FFF', fontSize: 9 },
                    viewWindowMode: 'explicit',
                },
                hAxis: {
                    gridlines: {
                        color: 'transparent'
                    },
                    textStyle: { color: '#FFF', fontSize: 9 }
                }
            };

            function calculateOverlaysBasedOnPredictionData(prediction, options) {
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

                const deadlineStr = prediction.summary.endedAtTruncatedDueToResultInvalidation ? prediction.summary.endedAtTruncatedDueToResultInvalidation : prediction.summary.endedAt
                const deadlineTs = deadlineStr ? new Date(deadlineStr).getTime() / 1000 : null
                if (deadlineTs && prediction.state.status === 'FINISHED') {
                    endedAtLineAt = new Date(deadlineStr)
                }

                const postedAtTs = new Date(prediction.postedAt).getTime() / 1000
                const firstCoinCandles = prediction.summary.candlestickMap[Object.keys(prediction.summary.candlestickMap)[0]]
                if (postedAtTs && postedAtTs >= firstCoinCandles[0].t && postedAtTs <= firstCoinCandles[firstCoinCandles.length - 1].t) {
                    postedAtLineAt = new Date(prediction.postedAt)
                }

                return {
                    upperGreenLineAt,
                    lowerGreenLineAt,
                    upperRedLineAt,
                    lowerRedLineAt,
                    postedAtLineAt,
                    endedAtLineAt,
                }
            }

            const {
                upperGreenLineAt,
                lowerGreenLineAt,
                upperRedLineAt,
                lowerRedLineAt,
                postedAtLineAt,
                endedAtLineAt,
            } = calculateOverlaysBasedOnPredictionData(prediction, options)

            // Overlays are visual elements that enhance the chart with information about the prediction.
            // They show e.g. when the prediction starts, when it ends and the areas in which they become correct/incorrect/invalidated.
            function placeOverlays(dataTable) {
                const cli = this.getChartLayoutInterface();
                const chartArea = cli.getChartAreaBoundingBox();

                const xMin = postedAtLineAt ? cli.getXLocation(postedAtLineAt) : chartArea.left
                const xMax = endedAtLineAt ? cli.getXLocation(endedAtLineAt) : chartArea.width + xMin

                if (upperGreenLineAt) {
                    showGoalRange({
                        isLower: false,
                        softGoal: upperGreenLineAt[1],
                        hardGoal: upperGreenLineAt[0],
                        softGoalLineClass: 'yellowLine',
                        hardGoalLineClass: 'greenLine',
                        softGoalBoxClass: 'yellowBox',
                        hardGoalBoxClass: 'greenBox',
                        softGoalCaption: `${prediction.summary.errorMarginRatio * 100}% error margin`,
                        hardGoalCaption: 'Goal',
                        xMin,
                        xMax,
                    }, cli)
                }

                if (lowerGreenLineAt) {
                    showGoalRange({
                        isLower: true,
                        softGoal: lowerGreenLineAt[1],
                        hardGoal: lowerGreenLineAt[0],
                        softGoalLineClass: 'yellowLine',
                        hardGoalLineClass: 'greenLine',
                        softGoalBoxClass: 'yellowBox',
                        hardGoalBoxClass: 'greenBox',
                        softGoalCaption: `${prediction.summary.errorMarginRatio * 100}% error margin`,
                        hardGoalCaption: 'Goal',
                        xMin,
                        xMax,
                    }, cli)
                }

                if (upperRedLineAt) {
                    showGoalRange({
                        isLower: false,
                        softGoal: upperRedLineAt[1],
                        hardGoal: upperRedLineAt[0],
                        softGoalLineClass: 'redLine',
                        hardGoalLineClass: 'redLine',
                        softGoalBoxClass: 'redBox',
                        hardGoalBoxClass: 'redBox',
                        softGoalCaption: `${prediction.summary.errorMarginRatio * 100}% error margin`,
                        hardGoalCaption: 'Incorrect',
                        xMin,
                        xMax,
                    }, cli)
                }

                if (lowerRedLineAt) {
                    showGoalRange({
                        isLower: true,
                        softGoal: lowerRedLineAt[1],
                        hardGoal: lowerRedLineAt[0],
                        softGoalLineClass: 'redLine',
                        hardGoalLineClass: 'redLine',
                        softGoalBoxClass: 'redBox',
                        hardGoalBoxClass: 'redBox',
                        softGoalCaption: `${prediction.summary.errorMarginRatio * 100}% error margin`,
                        hardGoalCaption: 'Incorrect',
                        xMin,
                        xMax,
                    }, cli)
                }

                // Show "finish line" white line with checkered flag emoji
                if (endedAtLineAt) {
                    showVerticalLine({ labelClass: 'endedAt', xMillis: endedAtLineAt }, cli)
                }

                // Show "starting line" white line with eyes emoji
                if (postedAtLineAt) {
                    showVerticalLine({ labelClass: 'postedAt', xMillis: postedAtLineAt, }, cli)
                }
            };

            // The "goal range" overlay shows a red/green box plus a dashed line and a caption delimiting an area of
            // the chart in which, if the candlesticks reach it, the prediction would become correct/incorrect/invalidated.
            function showGoalRange(args, cli) {
                const { isLower, softGoal, hardGoal, softGoalLineClass, hardGoalLineClass, softGoalBoxClass, hardGoalBoxClass, softGoalCaption, hardGoalCaption, xMin, xMax } = args

                const chartArea = cli.getChartAreaBoundingBox();
                const ySoftGoal = cli.getYLocation(softGoal)
                const yHardGoal = cli.getYLocation(hardGoal)
                const showSoftGoal = Math.abs(ySoftGoal - yHardGoal) > 15

                if (showSoftGoal) {
                    // Show the softGoalLine
                    showOverlay({
                        classList: [softGoalLineClass],
                        style: {
                            left: `${xMin}px`,
                            width: `${xMax - xMin}px`,
                            top: `${ySoftGoal}px`,
                        }
                    })

                    // Show the softGoalBox
                    showOverlay({
                        classList: [softGoalBoxClass],
                        style: {
                            left: `${xMin}px`,
                            width: `${xMax - xMin}px`,
                            top: isLower ? `${ySoftGoal}px` : `${yHardGoal + 1}px`,
                            height: isLower ? `${yHardGoal - ySoftGoal}px` : `${ySoftGoal - yHardGoal + 1}px`,
                        }
                    })

                    // Show the softGoalCaption
                    if (softGoalCaption) {
                        const isThereNoEndAtLine = xMax == chartArea.width + xMin
                        const padding = isThereNoEndAtLine ? 0 : 15
                        const elem = showOverlay({
                            classList: ['goalCaption'],
                            caption: softGoalCaption,
                            style: {
                                left: `${chartArea.left + chartArea.width + 5 + padding}px`,
                                height: 'auto',
                                width: 'auto',
                            }
                        })
                        elem.style.top = `${ySoftGoal - (elem.clientHeight / 2) + 1}px`;
                    }
                }

                // Show the hardGoalLine
                showOverlay({
                    classList: [hardGoalLineClass],
                    style: {
                        left: `${xMin}px`,
                        width: `${xMax - xMin}px`,
                        top: `${yHardGoal}px`,
                    }
                })

                // Show the hardGoalBox
                showOverlay({
                    classList: [hardGoalBoxClass],
                    style: {
                        left: `${xMin}px`,
                        width: `${xMax - xMin}px`,
                        top: isLower ? `${yHardGoal}px` : `${chartArea.top}px`,
                        height: isLower ? `${chartArea.height + chartArea.top - yHardGoal}px` : `${yHardGoal - chartArea.top}px`,
                    }
                })

                // Show the hardGoalCaption
                if (hardGoalCaption) {
                    const isThereNoEndAtLine = xMax == chartArea.width + xMin
                    const padding = isThereNoEndAtLine ? 0 : 15
                    const elem = showOverlay({
                        classList: ['goalCaption'],
                        caption: hardGoalCaption,
                        style: {
                            left: `${chartArea.left + chartArea.width + 5 + padding}px`,
                            height: 'auto',
                            width: 'auto',
                        }
                    })
                    elem.style.top = `${Math.round(yHardGoal - (elem.clientHeight / 2) + 1)}px`;
                    console.log({ top: elem.style.top, padding, left: chartArea.left + chartArea.width + 5 + padding })
                }
            }

            // The "vertical line" overlay is the white line with an icon on top, marking the start and end of a prediction.
            function showVerticalLine(args, cli) {
                const { lineClass, labelClass, xMillis } = args

                const chartArea = cli.getChartAreaBoundingBox();
                const x = cli.getXLocation(xMillis)

                // Show the vertical line
                showOverlay({
                    classList: ['verticalLine', lineClass],
                    style: {
                        left: `${x}px`,
                        top: `${chartArea.top}px`,
                        height: `${chartArea.height}px`,
                    }
                })

                // Show the icon on top of it
                showOverlay({
                    elemType: "img",
                    classList: [labelClass],
                    style: {
                        left: `${x - (20 / 2) + 1}px`,
                        top: `${chartArea.top - 20}px`,
                    }
                })
            }

            // All chart overlays are rendered into the DOM by this function.
            function showOverlay(args) {
                let { elemType, classList, style, caption } = args

                const elem = document.createElement(elemType ? elemType : 'div')
                classList.forEach(cls => elem.classList.add(cls))
                elem.classList.add('overlay')

                // Show before setting properties, so that clientHeight & clientWidth can be set
                document.querySelector('.chartWithMarkerOverlay').appendChild(elem)

                // Set caption first, so that it affects clientHeight & clientWidth
                if (caption) {
                    elem.innerText = caption
                }

                Object.entries(style).forEach(entry => {
                    const [key, value] = entry;
                    elem.style[key] = value
                });

                return elem // element is returned in case its client dimensions have to be used after rendering.
            }

            // Render the chart and place overlays once chart is ready
            var chart = new google.visualization.CandlestickChart(document.getElementById('line-chart-marker'));
            google.visualization.events.addListener(chart, 'ready',
                placeOverlays.bind(chart, data)
            );
            chart.draw(data, options);
        }
    </script>
</head>

<body>
    <div class="topSection">
        <div class="topLeft">
            <img class="profileImage" src="" />
            <div class="bubble bubble-bottom-left"></div>
            <span class="postAuthor"></span>
        </div>
        <div class="topRight">
            <span class="filler"></span>
            <span class="predictionResult"></span>
            <img id="predictionResultImage" />
        </div>
    </div>
    <div class="chartWithMarkerOverlay">
        <div id="line-chart-marker"></div>
    </div>
    <div class="footnote">* Goal may be reached at latest incomplete
        candlestick
        and not shown. Tweets don't store timezone information, which makes "end of day" and the like slightly
        innacurate.</div>
    <script type="text/javascript">
        document.querySelector('.profileImage').setAttribute("src", account.thumbnails[account.thumbnails.length - 1]);
        document.querySelector('.profileImage').setAttribute("onerror", "if (this.src != 'default_account_image.png') this.src = 'default_account_image.png'");
        document.querySelector('.bubble').innerText = prediction.predictionText;

        if (prediction.state.status !== 'FINISHED') {
            document.querySelector('.predictionResult').innerHTML = 'NOW TRACKING';
            document.querySelector('#predictionResultImage').src = 'eyes_1f440.png';
        } else if (prediction.state.value === 'CORRECT') {
            document.querySelector('.predictionResult').innerHTML = 'CORRECT';
            document.querySelector('#predictionResultImage').src = 'check-mark-button_2705.png';
        } else if (prediction.state.value === 'INCORRECT') {
            document.querySelector('.predictionResult').innerHTML = 'INCORRECT';
            document.querySelector('#predictionResultImage').src = 'cross-mark_274c.png';
        } else {
            document.querySelector('.predictionResult').innerHTML = 'INVALIDATED';
        }

        document.querySelector('.postAuthor').innerText = account.handle ? `@${account.handle}` : account.name;
    </script>

</body>

</html>
