<?php

namespace {
	// composer.json
	// {
	// 	"require": {
	// 		"php": ">=5.3.3",
	// 		"zendframework/zend-http": "2.2.4",
	// 		"zendframework/zend-uri": "2.2.4"
	// 	}
	// }
	if (file_exists(__DIR__ . "/../vendor/autoload.php")) require __DIR__ . "/../vendor/autoload.php";
}

namespace bbb {

	class BuddyClient {

		static public $ServerUrl = "http://localhost:8080/uh";
		private $client;

		public function __construct($url, $secret) {
			$this->client = new \Zend\Http\Client(self::$ServerUrl);
			$this->client->getRequest()->setMethod(\Zend\Http\Request::METHOD_POST);
			$this->client->setParameterGet(compact("url", "secret"));
		}

		public function setServerUrl($url) {
			$this->client->setUri($url);
			return $this;
		}

		private static function message($event, array $data = array()) {
			return json_encode(array("event" => $event, "data" => (object) $data));
		}

		public function create($id, array $options = array()) {
			return $this->failOn("create.fail", $this->emit("create", compact("id")+$options))->data;
		}

		public function joinUrl($name, $id, $password, array $options = array()) {
			return $this->emit("joinURL", compact("name", "id", "password")+$options)->data->url;
		}

		public function end($id, $password) {
			return $this->emit("end", compact("id", "password"))->data->ended;
		}

		public function isMeetingRunning($id) {
			return $this->emit("running", compact("id"))->data->running;
		}

		public function getMeetingInfo($id, $password) {
			return $this->failOn("info.fail", $this->emit("info", compact("id", "password")))->data;
		}

		public function getMeetings() {
			return $this->emit("meetings")->data->meetings;
		}

		public function getRecordings(array $meetings = array()) {
			return $this->emit("recordings", compact("meetings"))->data->recordings;
		}

		public function publishRecordings(array $recordings, $publish = true) {
			return $this->emit("recordings.publish", compact("recordings", "publish"))->data;
		}

		public function deleteRecordings(array $recordings) {
			return $this->emit("recordings.delete", compact("recordings"))->data;
		}

		private function emit($event, array $data = array()) {
			$this->client->setRawBody(self::message($event, $data));
			$response = $this->client->send();
			if (200 != $response->getStatusCode()) {
				throw new \Exception($response->getBody(), $response->getStatusCode());
			}
			return json_decode($response->getBody());
		}

		private function failOn($fail, \stdclass $event) {
			if ($fail == $event->event) throw new \Exception($event->data->error);
			return $event;
		}
	}
}

namespace {

	// \bbb\BuddyClient::$ServerUrl = "https://bbb.example.com/uh";
	// \bbb\BuddyClient::$ServerUrl = "http://localhost:8081/uh";

	$b4 = new \bbb\BuddyClient(
		"https://my.big.blue.button.com/bigbluebutton/api/",
		"5ea96baab0fabfab0deadc94197fd185"
	);

	$meeting = $b4->create(uniqid(), array(
		"logoutURL" => "http://localhost:8081/",
		"name" => "This meeting has NO name!",
		"welcome" => "Hi.",
	));

	var_dump(
		$meeting,
		$b4->getMeetingInfo($meeting->id, $meeting->moderatorPW),
		$b4->getMeetings()
	);

	echo "JoinURL: ", $b4->joinUrl(
		"Attendee ".uniqid(),
		$meeting->id,
		$meeting->attendeePW
	), PHP_EOL;

	var_dump($b4->getRecordings());
}

// object(stdClass)#9 (5) {
//   ["attendeePW"]=>
//   string(8) "Ktod567K"
//   ["created"]=>
//   int(1380900973)
//   ["forcedEnd"]=>
//   bool(false)
//   ["id"]=>
//   string(13) "524ee076d2ae5"
//   ["moderatorPW"]=>
//   string(8) "zOO7DlJQ"
// }
// object(stdClass)#11 (13) {
//   ["attendeePW"]=>
//   string(8) "Ktod567K"
//   ["created"]=>
//   int(1380900973)
//   ["endTime"]=>
//   int(0)
//   ["forcedEnd"]=>
//   bool(false)
//   ["id"]=>
//   string(13) "524ee076d2ae5"
//   ["maxUsers"]=>
//   int(0)
//   ["moderatorPW"]=>
//   string(8) "zOO7DlJQ"
//   ["name"]=>
//   string(25) "This meeting has NO name!"
//   ["numMod"]=>
//   int(0)
//   ["numUsers"]=>
//   int(0)
//   ["recording"]=>
//   bool(false)
//   ["running"]=>
//   bool(false)
//   ["stratTime"]=>
//   int(0)
// }
// array(3) {
//   [0]=>
//   object(stdClass)#19 (5) {
//     ["attendeePW"]=>
//     string(8) "Ktod567K"
//     ["created"]=>
//     int(1380900973)
//     ["forcedEnd"]=>
//     bool(false)
//     ["id"]=>
//     string(13) "524ee076d2ae5"
//     ["moderatorPW"]=>
//     string(8) "zOO7DlJQ"
//   }
//   [1]=>
//   object(stdClass)#14 (5) {
//     ["attendeePW"]=>
//     string(8) "UeVycYLQ"
//     ["created"]=>
//     int(1380900810)
//     ["forcedEnd"]=>
//     bool(false)
//     ["id"]=>
//     string(13) "524edfd327b48"
//     ["moderatorPW"]=>
//     string(8) "lWvrToKC"
//   }
//   [2]=>
//   object(stdClass)#17 (5) {
//     ["attendeePW"]=>
//     string(8) "hj2P2jcx"
//     ["created"]=>
//     int(1380900935)
//     ["forcedEnd"]=>
//     bool(false)
//     ["id"]=>
//     string(13) "524ee050831ed"
//     ["moderatorPW"]=>
//     string(8) "s8R8teiJ"
//   }
// }
// JoinURL: https://my.big.blue.button.com/bigbluebutton/api/join?checksum=9060a97ac439e64c5f1fe84b2d12992caf1baf49&fullName=Attendee+524ee0775daa4&meetingID=524ee076d2ae5&password=Ktod567K
// array(1) {
//   [0]=>
//   object(stdClass)#17 (6) {
//     ["endTime"]=>
//     int(1381314126745)
//     ["meetingId"]=>
//     string(0) ""
//     ["name"]=>
//     string(9) "Meeting 1"
//     ["playback"]=>
//     object(stdClass)#14 (3) {
//       ["len"]=>
//       int(3)
//       ["type"]=>
//       string(12) "presentation"
//       ["url"]=>
//       string(122) "http://my.big.blue.button.com/playback/presentation/playback.html?meetingId=09672563b912fc7919d14513124bb962fd6a42d7-1381313971571"
//     }
//     ["recordId"]=>
//     string(0) ""
//     ["startTime"]=>
//     int(1381313999223)
//   }
// }