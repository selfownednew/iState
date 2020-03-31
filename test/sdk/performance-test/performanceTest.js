var cluster = require('cluster');
var {
    sendreg,
    sendSimpleStore,
    createEMR,
    connect,
    deletereg,
} = require('./functions')

// Main
var sendfunc;
var startTPS;
var timeLimit; // In ms - 0 is infinity
var transactionLoadPerWorker; 
var appStartTime;
var highestSuccessTPS;
var lowestFailTPS; 
var currentTPS;
var numWorkers;
var latency;
var numSuccess;
var numFail;
var iteration;
var lastWorkerId;
var testStartTime;
var returnMsg;
var transactionSpeed;
var preConnectSocket;
var mainTransactionSpeed; // Autoset by function Dont change
var transactionSpeedRangeDelta; 
var randomTraffic;
var alwaysNewData;
var maxUniqueData;

function defineVariables() {
    alwaysNewData = false;
    maxUniqueData = 2000;
    randomTraffic = false;
    preConnectSocket = false;
    numWorkers = 1;
    transactionSpeed = 200; // 200 is the max accurate maybe for single worker
    transactionSpeedRangeDelta = 0; // Vary from x - 50 to x + 50 
    sendfunc = sendreg;
    startTPS = 10;
    timeLimit = 0; // In ms - 0 is infinity
    transactionLoadPerWorker = 100; 
    appStartTime = Date.now();
    highestSuccessTPS = -1;
    lowestFailTPS = 1000000000000; // 1 trillion TPS here considered practically impossible
    currentTPS = undefined;
    latency = undefined;
    numSuccess = undefined;
    numFail = undefined;
    iteration = 0;
    // recursiveReturnHandler = 0;
    lastWorkerId = 0;
    testStartTime = Date.now();
    returnMsg = `Test success`;
    latencyList = [];
}

function connectToPersistSockets(start, end, host, port){
    for (var i = start; i < end; i++) {
        connect(host, port);
    }
}

function randomNum(min, max) {
    return Math.floor(Math.random() * (max - min + 1) + min);
}

async function loadGenerator(TPS, start, end, host, port){
    console.log(`Each Load: ${TPS}`);
    console.log(`Start: ${start}, End: ${end}`);
    var startTime = Date.now();
    mainTransactionSpeed = transactionSpeed;
    for (var i = start; i < end; i++) {
        // console.log(`Request: ${i} sent.`);
        if (alwaysNewData) {
            sendfunc(i, TPS, cluster, host, port);
        } else { 
            sendfunc(i % maxUniqueData, TPS, cluster, host, port);
        }
        //await sleep(1000/TPS); // total TPS / 50 TPS
        if (randomTraffic) {
            mainTransactionSpeed = randomNum(transactionSpeed - transactionSpeedRangeDelta, transactionSpeed + transactionSpeedRangeDelta)
        }
        if (transactionSpeed != 0 ) {
            await sleep(1000/mainTransactionSpeed); // This for continuously pushing at same rate
            // console.log("SleepTime: ", 1000/mainTransactionSpeed);
            // console.log("Transaction Speed:", mainTransactionSpeed)
        }
    }
    var endTime = Date.now();
    var elapsedSecs = (endTime - startTime)/1000;
    console.log(`BatchLoopTime: `, elapsedSecs);

    return endTime;
    //do something when app is closing
    process.on('exit', function() {
        console.log(`Time: ${startTime/1000} - ${endTime/1000}`);
        console.log(`Seconds used to complete send request loop: ${elapsedSecs}`);
        console.log(`All completed in: ${(Date.now() - endTime)/1000}`);
    });

    // //catches ctrl+c event
    // process.on('SIGINT', exitHandler.bind(null, {exit:true}));

    // // catches "kill pid" (for example: nodemon restart)
    // process.on('SIGUSR1', exitHandler.bind(null, {exit:true}));
    // process.on('SIGUSR2', exitHandler.bind(null, {exit:true}));    
}

async function main(callback, host, port, maxTPS) {   
    if(!numWorkers) {
        numWorkers = Math.floor(currentTPS / transactionLoadPerWorker) || 1;
    }
    lastWorkerId = lastWorkerId + numWorkers;
    // numWorkers = 1;
    numSuccess = 0;
    numFail = 0;
    var tempLatency;
    var requestEndTime;
    var testLatency = 0;
    var requestStartTime;
    var requestEndTime;
    var firstWorkerStartTime = 0;
    var lastWorkerStartTime = 0;
    latencyList = [];
    if (cluster.isMaster){
        // var agent = new http.Agent({
        //     keepAlive: true,
        //     maxSockets: 1000,
        //     keepAliveMsecs: 3000
        // });
        iteration ++;
        console.log(`===================================================================================================`);
        console.log(`Iteration: ${iteration}`);
        console.log(`Testing TPS: ${currentTPS}`);
        console.log(`Number of workers: ${numWorkers}`);
        console.log(`Current highest TPS: ${highestSuccessTPS}`);
        console.log(`===================================================================================================`);
        for(var i=0; i<numWorkers; i++){        
            var new_worker_env = {};
            new_worker_env["currentTPS"] = currentTPS;
            new_worker_env["numWorkers"] = numWorkers;
            new_worker_env["lastWorkerId"] = lastWorkerId;
            new_worker_env["workerQue"] = Date.now() + 7000;
            var worker = cluster.fork(new_worker_env);
            requestStartTime = Date.now();
            worker.on('message', async function(msg) {
                // // HTTP
                // if(msg.getHTTPAgent){
                //     worker.send({
                //         agent: agent
                //     })
                // }
                // //
                // console.log(msg);
                if(msg.responseStat) {
                    numSuccess += msg.numSuccess;
                    numFail += msg.numFail;
                    console.log(`Worker - Success : Fail = ${msg.numSuccess} : ${msg.numFail}`);
                    console.log(`latency: ${msg.latency}`);
                    testLatency = testLatency + msg.latency;
                } else if(msg.timeStat) {
                   requestEndTime = msg.requestEndTime;
                    if (firstWorkerStartTime == 0) {
                        firstWorkerStartTime = msg.workerStartTime;
                    } else {
                        lastWorkerStartTime = msg.workerStartTime;
                    }
                }
                if(numSuccess + numFail === currentTPS) {
                    //tempLatency = (Date.now() - requestEndTime) / 1000;
                    tempLatency = testLatency / numWorkers;
                    await sleep(1000);
                    cluster.disconnect();
                    console.log(`TOTAL - Success : Fail = ${numSuccess} : ${numFail}`);
                    var cont = calculateNewTPS(numSuccess, numFail, tempLatency);
                    if(cont) {
                        if((Date.now() - appStartTime) < timeLimit || timeLimit === 0){ // Test timeout
                            if(currentTPS <= maxTPS) {
                                main(callback, host, port, maxTPS);
                            } else if(highestSuccessTPS < maxTPS) {
                                currentTPS = maxTPS;
                                main(callback, host, port, maxTPS);
                            } else {
                                console.log(`Stopping test due to TPS limit.`);
                                returnMsg = `Test stopped due to TPS-limit: ${maxTPS} TPS`
                                // process.exit();
                                returnFunc();
                            }
                        } else {
                            console.log(`Stopping test due to timelimit.`);
                            returnMsg = `Test stopped due to time-limit: ${timeLimit/1000}s`
                            // process.exit();
                            returnFunc();
                        }
                    } else {
                        // console.log(`CALLING EXIT...............`);
                        // process.exit();
                        returnFunc();
                    }

                    function returnFunc(){
                        // recursiveReturnHandler ++;
                        // console.log(`Inside return func: ${recursiveReturnHandler}, ${iteration}`)
                        // if(recursiveReturnHandler === iteration) {
                            console.log(`===================================================================================================`);
                            console.log(`RESULTS`);
                            console.log(`===================================================================================================`);
                            console.log(`Iteration: ${iteration}`);
                            console.log(`Highest TPS: ${highestSuccessTPS}`);
                            console.log(`Latency: ${latency}`);
                            console.log(`Difference in workers start time: ${(lastWorkerStartTime - firstWorkerStartTime) / 1000}`)
                            console.log(`Total test time taken (sec): ${(Date.now() - testStartTime)/1000}`);
                            console.log(`===================================================================================================`);
                            // return {
                            //     success: true,
                            //     msg: returnMsg, 
                            //     iteration,
                            //     highestSuccessTPS,
                            //     latency,
                            //     totalTestTime: ( Date.now() - testStartTime ) / 1000,
                            // }
                            callback({
                                success: true,
                                msg: returnMsg, 
                                iteration,
                                highestSuccessTPS,
                                latency,
                                totalTestTime: ( Date.now() - testStartTime ) / 1000,
                            });
                        // }
                    }
                }       
            });
        }
    
        cluster.on('exit', function(worker){
            // console.log(`${worker.id} - exited.`);
        });
        
        // process.on('exit', function() {
        //     recursiveReturnHandler ++;
        //     if(recursiveReturnHandler === iteration) {
        //         console.log(`===================================================================================================`);
        //         console.log(`RESULTS`);
        //         console.log(`===================================================================================================`);
        //         console.log(`Iteration: ${iteration}`);
        //         console.log(`Highest TPS: ${highestSuccessTPS}`);
        //         console.log(`Latency: ${latency}`);
        //         console.log(`Total test time taken (sec): ${(Date.now() - testStartTime)/1000}`);
        //         console.log(`===================================================================================================`);
        //         // return {
        //         //     success: true,
        //         //     msg: returnMsg, 
        //         //     iteration,
        //         //     highestSuccessTPS,
        //         //     latency,
        //         //     totalTestTime: ( Date.now() - testStartTime ) / 1000,
        //         // }
        //         callback({
        //             success: true,
        //             msg: returnMsg, 
        //             iteration,
        //             highestSuccessTPS,
        //             latency,
        //             totalTestTime: ( Date.now() - testStartTime ) / 1000,
        //         });


        //     }
        // });
        
    } else {
        //==========================================================================================
        // Http agent sharing with workers
        // var agent;
        // cluster.worker.send({
        //     getHTTPAgent: true
        // });
        // process.on('message', function(msg) {
        //     agent = msg.agent;
        //     console.log('AGENNNNNTTTT::::::::::::', agent);
        //     proceedWorker();
        // });
        // =========================================================================================
        // async function proceedWorker() {
            var workerStartTime = Date.now();
            var currentTPSfromWorker = process.env['currentTPS'];
            var updatedNumWorkers = process.env['numWorkers'];
            var updatedlastWorkerId = parseInt(process.env['lastWorkerId']);
            // console.log(`CURRENT TPS FROM WORKERRRR: ${currentTPSfromWorker}`);
            var eachLoad = Math.floor(currentTPSfromWorker / updatedNumWorkers);
            var remainder = currentTPSfromWorker % updatedNumWorkers;
            var start = eachLoad * (cluster.worker.id - 1);
            var end = eachLoad * cluster.worker.id;
            var workerQue = parseInt(process.env['workerQue']);
            // console.log(`Worker Id: ${cluster.worker.id}`); //// TO-DO FLAW : REMAINDER based on last worker id fail
            //console.log(`For ${cluster.worker.id}, Lastworkerid: ${updatedlastWorkerId}`);
            if(cluster.worker.id === updatedlastWorkerId) {
                //console.log(`I am the last worker: ${cluster.worker.id}`);
                eachLoad = eachLoad + remainder;
                end = end + remainder;
            }
            if(preConnectSocket) {
                connectToPersistSockets(start, end, host, port);
            }
            for(;;){
                if(Date.now() > workerQue) {
                    break;
                }
                await sleep(1);
            }
            var requestEndTime = await loadGenerator(eachLoad, start, end, host, port);
            cluster.worker.send({
                timeStat: true,
                requestEndTime,
                workerStartTime
            });
            //cluster.worker.send(`Worker: ${cluster.worker.id}`);
            // cluster.worker.disconnect();
        // }
    }   
}


function calculateNewTPS(numSuccess, numFail, tempLatency) {
    // console.log(`CALCULATING TPS.....................`);
    var cont = false;
    if(numFail === 0) { // If all success
        if(currentTPS > highestSuccessTPS) {
            highestSuccessTPS = currentTPS;
            latency = tempLatency;
        }
        if((currentTPS*2) < lowestFailTPS) { // If 2x current > failed TPS
            currentTPS *= 2;
        } else {
            currentTPS = Math.floor((currentTPS + lowestFailTPS) / 2); // Between current and lowest fail TPS
        }
        cont = true;
    } else { // If some fail
        if(currentTPS < lowestFailTPS) {
            lowestFailTPS = currentTPS;
        }
        if(numSuccess > highestSuccessTPS) {
            highestSuccessTPS = numSuccess;
            latency = tempLatency;
            // currentTPS = numSuccess;
            currentTPS = Math.floor((highestSuccessTPS + currentTPS) / 2);
            cont = true;
        } 
        // else {
        //     // Exit condition
        //     // cont = false , so do nothing.
        // }
    }
    // console.log(`CONTINUE = ${cont}.....................`);
    // console.log(`Current Highest TPS: ${highestSuccessTPS}`);
    return cont;
}

var beginTest = async function(inputTimeLimit, inputStartTPS, maxTPS, callback, host, port) {
    defineVariables();
    //  console.log(`Inside beginTest`);
    timeLimit = inputTimeLimit || timeLimit; 
    startTPS = inputStartTPS || startTPS;
    currentTPS = startTPS;
    // console.log(`${timeLimit} : ${startTPS} : ${currentTPS} : ${maxTPS} : ${host} : ${port} : ${callback}`);
    // masterId = cluster.worker.id;
    main(callback, host, port, maxTPS);
}
// main();
// Utils

// var customLoadgen = async function(rate, numReq, host, port) {
//     loadGenerator(rate, 0, numReq, host, port);
// }
function sleep(ms) {
	return new Promise(resolve => setTimeout(resolve, ms));
}

module.exports = {
    beginTest,
    // customLoadgen,
}