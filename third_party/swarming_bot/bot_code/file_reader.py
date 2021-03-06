# Copyright 2016 The LUCI Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

import json
import logging
import queue
import threading


class FatalReadError(Exception):
  """Raised by 'FileReaderThread.start' if it can't read the file."""


class FileReaderThread(object):
  """Represents a thread that periodically rereads file from disk.

  Used by task_runner to read authentication headers generated by bot_main.

  Uses JSON for serialization.

  The instance is not reusable (i.e. once stopped, cannot be started again).
  """

  def __init__(self, path, interval_sec=15, max_attempts=100):
    self._path = path
    self._interval_sec = interval_sec
    self._max_attempts = max_attempts
    self._thread = None
    self._signal = queue.Queue()
    self._lock = threading.Lock()
    self._last_value = None

  def start(self):
    """Starts the thread that periodically rereads the value.

    Once 'start' returns, 'last_value' can be used to grab the read value. It
    will be kept in-sync with the contents of the file until 'stop' is called.

    Raises:
      FatalReadError is the file cannot be read even after many retries.
    """
    assert self._thread is None
    self._read()  # initial read
    self._thread = threading.Thread(
        target=self._run, name='FileReaderThread %s' % self._path)
    self._thread.daemon = True
    self._thread.start()

  def stop(self):
    """Stops the reading thread (if it is running)."""
    if not self._thread:
      return
    self._signal.put(None)
    self._thread.join(60)  # don't wait forever
    if self._thread.is_alive():
      logging.error('FileReaderThread failed to terminate in time')

  @property
  def last_value(self):
    """Last read value."""
    with self._lock:
      return self._last_value

  def _read(self):
    """Attempts to read the file, retrying a bunch of times.

    Returns:
      True to carry on, False to exit the thread.
    """
    attempts = self._max_attempts
    while True:
      try:
        with open(self._path, 'rb') as f:
          body = json.load(f)
        with self._lock:
          if self._last_value != body:
            logging.info('Read %s', self._path)
            self._last_value = body
        return True  # success!
      except (IOError, OSError, ValueError) as e:
        last_error = 'Failed to read auth headers from %s: %s' % (self._path, e)
      attempts -= 1
      if not attempts:
        raise FatalReadError(last_error)
      if not self._wait(0.05):
        return False

  def _wait(self, timeout):
    """Waits for the given duration or until the stop signal.

    Returns:
      True if waited, False if received the stop signal.
    """
    try:
      self._signal.get(timeout=timeout)
      return False
    except queue.Empty:
      return True

  def _run(self):
    while self._wait(self._interval_sec):
      try:
        if not self._read():
          return
      except FatalReadError as e:
        # Log the error and simply keep last read value as it was. 'start'
        # makes sure to read it at least once.
        logging.error('Can\'t reread the file: %s', e)
