#!/usr/bin/env vpython3
# Copyright 2016 The LUCI Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

import logging
import os
import sys
import unittest

ROOT_DIR = os.path.dirname(
    os.path.abspath(os.path.join(__file__, os.pardir, os.pardir)))
sys.path.insert(0, ROOT_DIR)

from libs.luci_context import luci_context


@unittest.skip("crbug.com/1206509: the tests fail on CQ.")
class TestLuciContext(unittest.TestCase):
  def setUp(self):
    self.ek = luci_context.ENV_KEY
    # Makes all logged messages go into unittest's buffer to be revealed on test
    # failure.
    logging.root.handlers[0].stream = sys.stdout

  def tearDown(self):
    self.assertNotIn(
      self.ek, os.environ, '%s in environ (%r)! Possible leak in test?' % (
        self.ek, os.environ.get(self.ek)))
    luci_context._CUR_CONTEXT = None

  def test_ok(self):
    self.assertFalse(luci_context._check_ok('hi'))
    self.assertFalse(luci_context._check_ok({'hi': 'there'}))
    self.assertFalse(luci_context._check_ok({'hi': 'there',
                                             'ok': {'thing': 'true'}}))
    self.assertTrue(luci_context._check_ok({'ok': {'thing': 'true'}}))

  def test_initial_load_dne(self):
    self.assertDictEqual(luci_context.read_full(), {})
    self.assertDictEqual(luci_context._CUR_CONTEXT, {})

    def nope():
      raise Exception('I SHOULD NOT BE CALLED')
    og_load = luci_context._initial_load
    luci_context._initial_load = nope
    try:
      self.assertIsNone(luci_context.read('section'))
    finally:
      luci_context._initial_load = og_load

  def test_initial_load_not_json(self):
    with luci_context._tf("not json", data_raw=True) as name:
      os.environ[self.ek] = name
      try:
        self.assertDictEqual(luci_context.read_full(), {})
        self.assertDictEqual(luci_context._CUR_CONTEXT, {})
      finally:
        del os.environ[self.ek]

  def test_initial_load_cannot_read(self):
    with luci_context._tf({'something': {'data': True}}) as name:
      os.chmod(name, 0)
      os.environ[self.ek] = name
      try:
        self.assertDictEqual(luci_context.read_full(), {})
        self.assertDictEqual(luci_context._CUR_CONTEXT, {})
      finally:
        del os.environ[self.ek]

  def test_initial_load_not_dict(self):
    with luci_context._tf('hi') as name:
      os.environ[self.ek] = name
      try:
        self.assertDictEqual(luci_context.read_full(), {})
        self.assertDictEqual(luci_context._CUR_CONTEXT, {})
      finally:
        del os.environ[self.ek]

  def test_initial_load_not_subsection_dict(self):
    with luci_context._tf({'something': 'string'}) as name:
      os.environ[self.ek] = name
      try:
        self.assertDictEqual(luci_context.read_full(), {})
        self.assertDictEqual(luci_context._CUR_CONTEXT, {})
      finally:
        del os.environ[self.ek]

  def test_initial_load_win(self):
    with luci_context.write(something={'data': True}):
      self.assertDictEqual(luci_context.read_full(),
                           {'something': {'data': True}})
      self.assertDictEqual(luci_context._CUR_CONTEXT,
                           {'something': {'data': True}})
      self.assertDictEqual(luci_context.read('something'), {'data': True})

  def test_nested(self):
    w = luci_context.write
    r = luci_context.read
    with w(something={'data': True}):
      self.assertIsNone(r('other'))
      self.assertDictEqual(r('something'), {'data': True})

      with w(other={'not': 10}, something=None):
        self.assertIsNone(r('something'))
        self.assertDictEqual(r('other'), {'not': 10})

      self.assertIsNone(r('other'))
      self.assertDictEqual(r('something'), {'data': True})

    self.assertIsNone(r('other'))
    self.assertIsNone(r('something'))

  def test_write_unchanged(self):
    with luci_context.write(something={'data': True}):
      path = os.environ.get(self.ek)
      with luci_context.write():
        self.assertDictEqual(luci_context.read_full(),
                             {'something': {'data': True}})
        self.assertEqual(os.environ.get(self.ek), path)
      with luci_context.write(something={'data': True}):
        self.assertEqual(os.environ.get(self.ek), path)
        self.assertDictEqual(luci_context.read_full(),
                             {'something': {'data': True}})
      with luci_context.write(something={'data': False}):
        self.assertNotEqual(os.environ.get(self.ek), path)

  def test_stage(self):
    path = None
    with luci_context.stage(something={'data': True}) as path:
      with open(path, 'r') as f:
        self.assertEqual('{"something": {"data": true}}', f.read())
    # The file is gone outside 'with' block.
    self.assertFalse(os.path.exists(path))

  def test_stage_unchanged(self):
    with luci_context.write(something={'data': True}):
      path = os.environ.get(self.ek)
      with luci_context.stage() as new_path:
        self.assertIsNone(new_path)
      with luci_context.stage(something={'data': True}) as new_path:
        self.assertIsNone(new_path)
      with luci_context.stage(something={'data': False}) as new_path:
        self.assertNotEqual(new_path, path)

  def test_leak(self):
    path = None
    with luci_context._tf({'something': {'data': True}}, leak=True) as path:
      self.assertTrue(os.path.exists(path))
    # The file is not deleted after contextmanager exits
    self.assertTrue(os.path.exists(path))
    os.unlink(path)


if __name__ == '__main__':
  # Pop it out of the environment to make sure we start clean.
  logging.basicConfig()
  os.environ.pop(luci_context.ENV_KEY, None)
  unittest.main(buffer=True)
