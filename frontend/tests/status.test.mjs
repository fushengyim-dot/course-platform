import {test} from 'node:test';import assert from 'node:assert/strict';import {classify,messages} from '../src/status.mjs';
test('S8 normal: successful healthy response',()=>assert.equal(classify(true,{code:'OK',data:{dependencies:{mysql:'up',qdrant:'up'}}}),'ready'));
test('S8 failure: HTTP or business or dependency failure',()=>{assert.equal(classify(false,{code:'OK'}),'error');assert.equal(classify(true,{code:'DEPENDENCY_UNAVAILABLE'}),'error');assert.equal(classify(true,{code:'OK',data:{dependencies:{mysql:'down'}}}),'error')});
test('S8 boundary: no dependency data is empty',()=>{assert.equal(classify(true,{code:'OK',data:null}),'empty');assert.equal(classify(true,{code:'OK',data:{dependencies:{}}}),'empty');assert.equal(Object.keys(messages).length,4)});
